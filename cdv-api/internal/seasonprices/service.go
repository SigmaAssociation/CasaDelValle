package seasonprices

import (
	"context"
	"errors"
)

type SeasonPriceRepository interface {
	GetByID(ctx context.Context, id int) (*SeasonPrice, error)
	Delete(ctx context.Context, id int) error
	GetCabinHostID(ctx context.Context, cabinID uint) (uint, error)
}

type Service struct {
	repository SeasonPriceRepository
}

func NewService(repository SeasonPriceRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetCabinHostID(ctx context.Context, cabinID uint) (uint, error) {
	if cabinID == 0 {
		return 0, errors.New("La cabaña es requerida")
	}

	return s.repository.GetCabinHostID(ctx, cabinID)
}

func (s *Service) Delete(ctx context.Context, cabinID uint, id int) error {
	if cabinID == 0 {
		return errors.New("La cabaña es requerida")
	}
	if id <= 0 {
		return errors.New("ID de tarifa inválido")
	}

	seasonPrice, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if seasonPrice.CabinID != cabinID {
		return ErrSeasonPriceNotFound
	}

	return s.repository.Delete(ctx, id)
}