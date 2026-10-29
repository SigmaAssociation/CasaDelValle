#!/bin/sh
set -eu

# La ejecución exterior usa siempre el Compose de desarrollo del repositorio.
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
if [ "${1:-}" != "--container" ]; then
    if [ "$#" -ne 0 ] && [ "$#" -ne 2 ]; then
        printf 'Uso: %s [USUARIO BASE_DE_DATOS]\n' "$0" >&2
        exit 1
    fi
    exec docker compose --project-directory "$script_dir/.." \
        -f "$script_dir/../docker-compose.yml" exec -T postgres-db \
        sh /migrations/migrate.sh --container "$@"
fi
shift
if [ "$#" -ne 0 ] && [ "$#" -ne 2 ]; then
    printf 'Se requieren usuario y base de datos juntos.\n' >&2
    exit 1
fi
db_user=${1:-${POSTGRES_USER:?}}
db_name=${2:-${POSTGRES_DB:?}}

for migration in "$script_dir"/migrations/*.sql; do
    [ -f "$migration" ] || continue
    name=${migration##*/}
    # Estos nombres se interpolan en SQL y en comandos de psql.
    case "$name" in
        *[!a-zA-Z0-9_.-]*) printf 'Nombre de migración inválido: %s\n' "$name" >&2; exit 1 ;;
    esac
done

# Una sola conexión mantiene el bloqueo durante toda la ejecución. El historial
# y los cambios de cada archivo se confirman juntos, o se revierten juntos.
{
    cat <<'SQL'
SELECT pg_advisory_lock(728194630);
CREATE TABLE IF NOT EXISTS public.schema_migrations (
    name TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
SQL
    for migration in "$script_dir"/migrations/*.sql; do
        [ -f "$migration" ] || continue
        name=${migration##*/}
        printf "SELECT EXISTS (SELECT 1 FROM public.schema_migrations WHERE name = '%s') AS applied \\gset\n" "$name"
        printf '\\if :applied\n\\echo Omitiendo %s (ya aplicada).\n\\else\n' "$name"
        printf '\\echo Aplicando %s...\nBEGIN;\n\\i %s\n' "$name" "$migration"
        printf "INSERT INTO public.schema_migrations (name) VALUES ('%s');\nCOMMIT;\n" "$name"
        printf '\\endif\n'
    done
    printf '\\echo Migraciones completadas.\n'
} | psql -X -v ON_ERROR_STOP=1 -U "$db_user" -d "$db_name"
