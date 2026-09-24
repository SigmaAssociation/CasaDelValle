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

func validateCabinName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 3 || len(name) > 150 {
		return errors.New("Nombre inválido: debe tener entre 3 y 150 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑüÜ\s\-\.\'#]+$`, name); !matched {
		return errors.New("Nombre inválido: contiene caracteres no permitidos")
	}
	return nil
}

func validateCabinAddress(address string) error {
	address = strings.TrimSpace(address)
	if len(address) < 5 || len(address) > 255 {
		return errors.New("Dirección inválida: debe tener entre 5 y 255 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$`, address); !matched {
		return errors.New("Dirección inválida: contiene caracteres no permitidos")
	}
	return nil
}

func validateCabinCommonFields(price float64, description string, capacity int, rules string) error {
	if price <= 0 {
		return errors.New("Precio inválido: debe ser mayor a 0")
	}

	if len(strings.TrimSpace(description)) > 500 {
		return errors.New("Descripción inválida: debe tener un máximo de 500 caracteres")
	}

	if capacity < 1 || capacity > 50 {
		return errors.New("Capacidad inválida: debe estar entre 1 y 50")
	}

	if len(strings.TrimSpace(rules)) > 500 {
		return errors.New("Reglas inválidas: deben tener un máximo de 500 caracteres")
	}

	return nil
}

func (s *Service) CreateCabin(ctx context.Context, req CreateCabinRequest) (int, error) {
	if err := validateCabinName(req.Name); err != nil {
		return 0, err
	}
	if err := validateCabinAddress(req.Address); err != nil {
		return 0, err
	}
	if err := validateCabinCommonFields(req.Price, req.Description, req.Capacity, req.Rules); err != nil {
		return 0, err
	}

	if req.HostID <= 0 {
		return 0, errors.New("Anfitrión inválido: es obligatorio indicar el anfitrión de la cabaña")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	req.Description = strings.TrimSpace(req.Description)
	req.Rules = strings.TrimSpace(req.Rules)

	return s.repository.CreateCabin(ctx, req)
}

func (s *Service) UpdateCabin(ctx context.Context, id int, req UpdateCabinRequest) error {
	if id <= 0 {
		return errors.New("ID de cabaña inválido")
	}
	if err := validateCabinName(req.Name); err != nil {
		return err
	}
	if err := validateCabinAddress(req.Address); err != nil {
		return err
	}
	if err := validateCabinCommonFields(req.Price, req.Description, req.Capacity, req.Rules); err != nil {
		return err
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	req.Description = strings.TrimSpace(req.Description)
	req.Rules = strings.TrimSpace(req.Rules)

	if err := s.repository.Update(ctx, id, req); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "Cabaña no encontrada" {
			return errors.New("Cabaña no encontrada")
		}
		return err
	}

	return nil
}

func (s *Service) DeleteCabin(ctx context.Context, id int) (int, error) {
	if id <= 0 {
		return 0, errors.New("ID de cabaña inválido")
	}

	rowsAffected, err := s.repository.DeleteCabin(ctx, id)
	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		return 0, errors.New("Cabaña no encontrada")
	}

	return rowsAffected, nil
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
	if params.Name != nil && strings.TrimSpace(*params.Name) == "" {
		params.Name = nil
	}
	if params.HostName != nil && strings.TrimSpace(*params.HostName) == "" {
		params.HostName = nil
	}
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
