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
	req.Address = strings.TrimSpace(req.Address)
	if len(req.Address) < 5 || len(req.Address) > 255 {
		return 0, errors.New("Dirección inválida: debe tener entre 5 y 255 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$`, req.Address); !matched {
		return 0, errors.New("Dirección inválida: contiene caracteres no permitidos")
	}

	if req.Price <= 0 {
		return 0, errors.New("Precio inválido: debe ser mayor a 0")
	}

	req.Description = strings.TrimSpace(req.Description)
	if len(req.Description) > 500 {
		return 0, errors.New("Descripción inválida: debe tener un máximo de 500 caracteres")
	}

	if req.Capacity < 1 || req.Capacity > 50 {
		return 0, errors.New("Capacidad inválida: debe estar entre 1 y 50")
	}

	req.Rules = strings.TrimSpace(req.Rules)
	if len(req.Rules) > 500 {
		return 0, errors.New("Reglas inválidas: deben tener un máximo de 500 caracteres")
	}

	if req.HostID <= 0 {
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

func (s *Service) UpdateCabin(ctx context.Context, req UpdateCabinRequest, userID int, userRole int) error {
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 3 || len(req.Name) > 150 {
		return errors.New("Nombre inválido: debe tener entre 3 y 150 caracteres")
	}
	req.Address = strings.TrimSpace(req.Address)
	if len(req.Address) < 5 || len(req.Address) > 255 {
		return errors.New("Dirección inválida: debe tener entre 5 y 255 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$`, req.Address); !matched {
		return errors.New("Dirección inválida: contiene caracteres no permitidos")
	}
	if req.Price <= 0 {
		return errors.New("Precio inválido: debe ser mayor a 0")
	}
	req.Description = strings.TrimSpace(req.Description)
	if len(req.Description) > 500 {
		return errors.New("Descripción inválida: debe tener un máximo de 500 caracteres")
	}
	if req.Capacity < 1 || req.Capacity > 50 {
		return errors.New("Capacidad inválida: debe estar entre 1 y 50")
	}
	req.Rules = strings.TrimSpace(req.Rules)
	if len(req.Rules) > 500 {
		return errors.New("Reglas inválidas: deben tener un máximo de 500 caracteres")
	}
	if userRole != 1 {
		hostID, err := s.repository.GetHostID(ctx, req.ID)
		if err != nil {
			return errors.New("la cabaña especificada no existe")
		}
		if hostID != userID {
			return errors.New("no tienes permisos para editar esta cabaña")
		}
	}

	return s.repository.Update(ctx, req)
}

func (s *Service) DeleteCabin(ctx context.Context, id int, userID int, userRole int) error {
	if userRole != 1 {
		hostID, err := s.repository.GetHostID(ctx, id)
		if err != nil {
			return errors.New("la cabaña especificada no existe")
		}
		if hostID != userID {
			return errors.New("no tienes permisos para eliminar esta cabaña")
		}
	}

	return s.repository.Delete(ctx, id)
}
