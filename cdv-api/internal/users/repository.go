package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
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

func (r *Repository) CreateUser(ctx context.Context, req RegisterRequest, hashPassword string) (int, error) {
	var id int
	query := "INSERT INTO usuarios (nombre, dpi, correo, contrasena, id_rol, direccion, telefono) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	err := r.pool.QueryRow(ctx, query, req.Name, req.DPI, req.Email, hashPassword, 2, req.Address, req.Phone).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return 0, errors.New("El correo o dpi ya está registrado")
		}
		return 0, err
	}
	return id, nil
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
			&user.Address,
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

func (r *Repository) GetUserAuthInfo(ctx context.Context, email string) (UserAuth, error) {
	var user UserAuth

	query := "SELECT id, nombre, correo, contrasena, id_rol FROM usuarios WHERE correo = $1"
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.IDRole,
	)

	if err != nil {
		return UserAuth{}, err
	}

	return user, nil
}
