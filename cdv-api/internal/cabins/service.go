package cabins

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetCabins(ctx context.Context) ([]Cabin, error) {
	return s.repository.GetAll(ctx)
}

func (s *Service) GetCabinByID(ctx context.Context, req GetCabinByIDRequest) (Cabin, error) {
	if req.ID <= 0 {
		return Cabin{}, errors.New("ID de cabaña inválido")
	}

	cabin, err := s.repository.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Cabin{}, errors.New("Cabaña no encontrada")
		}

		return Cabin{}, err
	}

	return cabin, nil
}

func (s *Service) GetCabinsByUser(ctx context.Context, req GetCabinsByUserRequest) ([]Cabin, error) {
	if req.UserID <= 0 {
		return nil, errors.New("ID de usuario inválido")
	}

	return s.repository.GetByHostID(ctx, req.UserID)
}
