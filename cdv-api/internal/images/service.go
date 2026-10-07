package images

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

// GetImagesByCabin retorna todas las imágenes de una cabaña.
func (s *Service) GetImagesByCabin(ctx context.Context, cabinID int) (ImagesListResponse, error) {
	if cabinID <= 0 {
		return ImagesListResponse{}, errors.New("ID de cabaña inválido")
	}

	images, err := s.repository.GetByCabinID(ctx, cabinID)
	if err != nil {
		return ImagesListResponse{}, err
	}

	return ToResponseList(images), nil
}

// GetMainImageByCabin retorna la imagen más reciente de una cabaña.
func (s *Service) GetMainImageByCabin(ctx context.Context, cabinID int) (Image, error) {
	if cabinID <= 0 {
		return Image{}, errors.New("ID de cabaña inválido")
	}

	image, err := s.repository.GetMainByCabinID(ctx, cabinID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Image{}, errors.New("La cabaña no tiene imágenes registradas")
		}

		return Image{}, err
	}

	return image, nil
}

// GetImageByID retorna una imagen por su identificador.
func (s *Service) GetImageByID(ctx context.Context, id int) (Image, error) {
	if id <= 0 {
		return Image{}, errors.New("ID de imagen inválido")
	}

	image, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Image{}, errors.New("Imagen no encontrada")
		}

		return Image{}, err
	}

	return image, nil
}
