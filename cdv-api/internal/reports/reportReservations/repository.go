package reportreservations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ObtenerReporte(ctx context.Context, filtro FiltroReporte) ([]CabanaReporte, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

const queryReporte = `
SELECT
    c.id,
    c.nombre,
    c.direccion,
    c.precio,
    c.capacidad,
    a.id,
    a.nombre,
    COUNT(r.id) OVER (PARTITION BY c.id) AS total_reservaciones,
    r.id,
    r.fecha_inicio,
    r.fecha_fin,
    r.estado,
    r.fecha_creacion,
    h.id,
    h.nombre
FROM cabanas c
JOIN usuarios a ON a.id = c.id_anfitrion
LEFT JOIN reservaciones r
       ON r.id_cabana = c.id
      AND r.estado <> 'cancelada'
LEFT JOIN usuarios h ON h.id = r.id_usuario
WHERE ($1::int IS NULL OR c.id_anfitrion = $1)
  AND ($2::int IS NULL OR c.id = $2)
ORDER BY total_reservaciones DESC, c.nombre ASC, c.id ASC, r.fecha_inicio DESC, r.id DESC
`

func (r *repository) ObtenerReporte(ctx context.Context, filtro FiltroReporte) ([]CabanaReporte, error) {
	rows, err := r.pool.Query(ctx, queryReporte, filtro.IDAnfitrion, filtro.IDCabana)
	if err != nil {
		return nil, fmt.Errorf("consultando reporte de reservaciones: %w", err)
	}
	defer rows.Close()

	cabanas := []CabanaReporte{}
	var idActual int

	for rows.Next() {
		var (
			cab CabanaReporte

			resID         *int
			resInicio     *time.Time
			resFin        *time.Time
			resEstado     *string
			resCreacion   *time.Time
			huespedID     *int
			huespedNombre *string
		)

		err := rows.Scan(
			&cab.ID,
			&cab.Nombre,
			&cab.Direccion,
			&cab.Precio,
			&cab.Capacidad,
			&cab.Anfitrion.ID,
			&cab.Anfitrion.Nombre,
			&cab.TotalReservaciones,
			&resID,
			&resInicio,
			&resFin,
			&resEstado,
			&resCreacion,
			&huespedID,
			&huespedNombre,
		)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila del reporte: %w", err)
		}

		if len(cabanas) == 0 || cab.ID != idActual {
			cab.Reservaciones = []ReservacionResumen{} 
			cabanas = append(cabanas, cab)
			idActual = cab.ID
		}

		if resID == nil {
			continue
		}

		actual := &cabanas[len(cabanas)-1]
		actual.Reservaciones = append(actual.Reservaciones, ReservacionResumen{
			ID:            *resID,
			FechaInicio:   *resInicio,
			FechaFin:      *resFin,
			Estado:        *resEstado,
			FechaCreacion: *resCreacion,
			Huesped: UsuarioResumen{
				ID:     *huespedID,
				Nombre: *huespedNombre,
			},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterando filas del reporte: %w", err)
	}

	return cabanas, nil
}