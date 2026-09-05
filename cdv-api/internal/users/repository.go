package users

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetAll(ctx context.Context) ([]User, error) {

	rows, err := r.pool.Query(
		ctx,
		`SELECT id, nombre, telefono, direccion, dpi, correo, id_rol
		 FROM usuarios
		 ORDER BY id`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var usuarios []User

	for rows.Next() {

		var user User

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Phone,
			&user.Adress,
			&user.DPI,
			&user.Email,
			&user.IDRole,
		)

		if err != nil {
			return nil, err
		}

		usuarios = append(usuarios, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return usuarios, nil
}
