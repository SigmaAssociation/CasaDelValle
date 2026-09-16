package cabins

import (
	"context"
	"errors"
	"regexp"
	"strings"

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

func (s *Service) CreateCabin(ctx context.Context, req CreateCabinRequest) (int, error) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	if len(req.Nombre) < 3 || len(req.Nombre) > 150 {
		return 0, errors.New("Nombre inválido: debe tener entre 3 y 150 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑüÜ\s\-\.\'#]+$`, req.Nombre); !matched {
		return 0, errors.New("Nombre inválido: contiene caracteres no permitidos")
	}

	req.Direccion = strings.TrimSpace(req.Direccion)
	if len(req.Direccion) < 5 || len(req.Direccion) > 255 {
		return 0, errors.New("Dirección inválida: debe tener entre 5 y 255 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$`, req.Direccion); !matched {
		return 0, errors.New("Dirección inválida: contiene caracteres no permitidos")
	}

	if req.Precio <= 0 {
		return 0, errors.New("Precio inválido: debe ser mayor a 0")
	}

	req.Descripcion = strings.TrimSpace(req.Descripcion)
	if len(req.Descripcion) > 500 {
		return 0, errors.New("Descripción inválida: debe tener un máximo de 500 caracteres")
	}

	if req.Capacidad < 1 || req.Capacidad > 50 {
		return 0, errors.New("Capacidad inválida: debe estar entre 1 y 50")
	}

	req.Reglas = strings.TrimSpace(req.Reglas)
	if len(req.Reglas) > 500 {
		return 0, errors.New("Reglas inválidas: deben tener un máximo de 500 caracteres")
	}

	if req.IDAnfitrion <= 0 {
		return 0, errors.New("Anfitrión inválido: es obligatorio indicar el anfitrión de la cabaña")
	}

	return s.repository.CreateCabin(ctx, req)
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

func (s *Service) SearchCabins(ctx context.Context, params CabinSearchParams) (CabinCardsListResponse, error) {
	if params.MinCapacity != nil && *params.MinCapacity < 0 {
		return CabinCardsListResponse{}, errors.New("la capacidad mínima no puede ser negativa")
	}
	if params.MaxCapacity != nil && *params.MaxCapacity < 0 {
		return CabinCardsListResponse{}, errors.New("la capacidad máxima no puede ser negativa")
	}
	if params.MinCapacity != nil && params.MaxCapacity != nil && *params.MinCapacity > *params.MaxCapacity {
		return CabinCardsListResponse{}, errors.New("la capacidad mínima no puede ser mayor que la máxima")
	}

	if params.MinPrice != nil && *params.MinPrice < 0 {
		return CabinCardsListResponse{}, errors.New("el precio mínimo no puede ser negativo")
	}
	if params.MaxPrice != nil && *params.MaxPrice < 0 {
		return CabinCardsListResponse{}, errors.New("el precio máximo no puede ser negativo")
	}
	if params.MinPrice != nil && params.MaxPrice != nil && *params.MinPrice > *params.MaxPrice {
		return CabinCardsListResponse{}, errors.New("el precio mínimo no puede ser mayor que el precio máximo")
	}

	if params.HostID != nil && *params.HostID <= 0 {
		return CabinCardsListResponse{}, errors.New("ID de anfitrión inválido")
	}

	cabins, err := s.repository.SearchCabins(ctx, params)
	if err != nil {
		return CabinCardsListResponse{}, err
	}

	return CabinCardsListResponse{
		Data:  cabins,
		Total: len(cabins),
	}, nil
}
