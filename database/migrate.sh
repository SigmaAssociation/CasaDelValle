#!/bin/sh

set -eu

migrations_dir="database/migrations"
completed_file="database/completed_migrations.txt"

if [ "$#" -ne 2 ]; then
    printf 'Uso: %s USUARIO BASE_DE_DATOS\n' "$0" >&2
    exit 1
fi

db_user="$1"
db_name="$2"

touch "$completed_file"

set -- "$migrations_dir"/*.sql
if [ "$1" = "$migrations_dir/*.sql" ]; then
    printf 'No hay migraciones por ejecutar.\n'
    exit 0
fi

for migration in "$@"; do
    migration_name="${migration##*/}"

    if grep -Fxq "$migration_name" "$completed_file"; then
        printf 'Omitiendo %s (ya aplicada).\n' "$migration_name"
        continue
    fi

    printf 'Aplicando %s...\n' "$migration_name"
    docker exec -i postgres_db psql -v ON_ERROR_STOP=1 -U "$db_user" -d "$db_name" < "$migration"
    printf '%s\n' "$migration_name" >> "$completed_file"
done

printf 'Migraciones completadas.\n'
