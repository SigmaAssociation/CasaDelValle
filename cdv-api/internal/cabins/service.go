package cabins

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	"cdv-api/internal/notifications"
)

type Service struct {
	repository CabinRepository
	notifier   Notifier
}

type Notifier interface {
	Notify(ctx context.Context, input notifications.NotificationInput)
	NotifyAdmins(ctx context.Context, input notifications.NotificationInput)
}

// CabinRepository abstrae el acceso a datos de cabañas para poder sustituirlo
// en las pruebas unitarias.
type CabinRepository interface {
	GetAll(ctx context.Context) ([]Cabin, error)
	CreateCabin(ctx context.Context, req CreateCabinRequest) (int, error)
	GetByID(ctx context.Context, id int) (Cabin, error)
	Update(ctx context.Context, id int, req UpdateCabinRequest) error
	DeleteCabin(ctx context.Context, id int) (int, error)
	GetByHostID(ctx context.Context, hostID int) ([]Cabin, error)
	SearchCabins(ctx context.Context, params CabinSearchParams) ([]CabinCardResponse, error)
}

func NewService(repository CabinRepository, notifier Notifier) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
	}
}

func (s *Service) notify(ctx context.Context, userID int, tipo, mensaje, entidadTipo string, entidadID int) {
	if s.notifier == nil || userID <= 0 {
		return
	}
	s.notifier.Notify(ctx, notifications.NotificationInput{
		UserID:      uint(userID),
		Tipo:        tipo,
		Mensaje:     mensaje,
		EntidadTipo: entidadTipo,
		EntidadID:   entidadID,
	})
}

// notifyAdmins replica un evento del sistema en el buzón de los administradores.
func (s *Service) notifyAdmins(ctx context.Context, tipo, mensaje string, entidadID int) {
	if s.notifier == nil {
		return
	}
	s.notifier.NotifyAdmins(ctx, notifications.NotificationInput{
		Tipo:        tipo,
		Mensaje:     mensaje,
		EntidadTipo: notifications.EntidadCabana,
		EntidadID:   entidadID,
	})
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

	cabinID, err := s.repository.CreateCabin(ctx, req)
	if err != nil {
		return 0, err
	}

	s.notify(ctx, req.HostID, notifications.TipoCabanaCreada,
		fmt.Sprintf("Tu cabaña %s fue creada correctamente.", req.Name),
		notifications.EntidadCabana, cabinID)

	s.notifyAdmins(ctx, notifications.TipoCabanaCreada,
		fmt.Sprintf("Se registró la cabaña %s (anfitrión #%d).", req.Name, req.HostID), cabinID)

	return cabinID, nil
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

	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("Cabaña no encontrada")
		}
		return err
	}

	if err := s.repository.Update(ctx, id, req); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "Cabaña no encontrada" {
			return errors.New("Cabaña no encontrada")
		}
		return err
	}

	s.notify(ctx, existing.HostID, notifications.TipoCabanaActualizada,
		fmt.Sprintf("Tu cabaña %s fue actualizada.", req.Name),
		notifications.EntidadCabana, id)

	s.notifyAdmins(ctx, notifications.TipoCabanaActualizada,
		fmt.Sprintf("Se actualizó la cabaña %s (anfitrión #%d).", req.Name, existing.HostID), id)

	return nil
}

func (s *Service) DeleteCabin(ctx context.Context, id int) (int, error) {
	if id <= 0 {
		return 0, errors.New("ID de cabaña inválido")
	}

	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("Cabaña no encontrada")
		}
		return 0, err
	}

	rowsAffected, err := s.repository.DeleteCabin(ctx, id)
	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		return 0, errors.New("Cabaña no encontrada")
	}

	s.notify(ctx, existing.HostID, notifications.TipoCabanaEliminada,
		fmt.Sprintf("Tu cabaña %s fue eliminada.", existing.Name),
		notifications.EntidadCabana, id)

	s.notifyAdmins(ctx, notifications.TipoCabanaEliminada,
		fmt.Sprintf("Se eliminó la cabaña %s (anfitrión #%d).", existing.Name, existing.HostID), id)

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
