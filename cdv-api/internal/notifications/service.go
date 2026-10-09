package notifications

import (
	"context"
	"errors"
	"log"
	"strings"
)

type Service struct {
	repository NotificationRepository
}

type NotificationRepository interface {
	Create(ctx context.Context, input NotificationInput) (int, error)
	GetByUser(ctx context.Context, userID uint, limit, offset int) ([]Notification, error)
	CountByUser(ctx context.Context, userID uint) (int, error)
	GetUnreadCount(ctx context.Context, userID uint) (int, error)
	GetAdminUserIDs(ctx context.Context) ([]uint, error)
	MarkAllRead(ctx context.Context, userID uint) (int, error)
	MarkRead(ctx context.Context, id int, userID uint) error
}

func NewService(repository NotificationRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func validateInput(input NotificationInput) error {
	if input.UserID == 0 {
		return ErrUserRequired
	}
	if strings.TrimSpace(input.Tipo) == "" {
		return errors.New("El tipo de notificación es requerido")
	}
	if len(input.Tipo) > MaxTipoLength {
		return errors.New("El tipo de notificación es demasiado largo")
	}
	if !IsValidTipo(input.Tipo) {
		return errors.New("Tipo de notificación no válido")
	}
	if strings.TrimSpace(input.Mensaje) == "" {
		return errors.New("El mensaje de la notificación es requerido")
	}
	if len(input.Mensaje) > MaxMensajeSize {
		return errors.New("El mensaje de la notificación es demasiado largo")
	}
	if !IsValidEntidad(input.EntidadTipo) {
		return errors.New("Tipo de entidad no válido")
	}
	if input.EntidadID < 0 {
		return errors.New("ID de entidad inválido")
	}
	return nil
}

func (s *Service) Create(ctx context.Context, input NotificationInput) (int, error) {
	if err := validateInput(input); err != nil {
		return 0, err
	}
	return s.repository.Create(ctx, input)
}

func (s *Service) Notify(ctx context.Context, input NotificationInput) {
	if _, err := s.Create(ctx, input); err != nil {
		log.Printf("notifications: no se pudo crear la notificación (usuario %d, tipo %s): %v",
			input.UserID, input.Tipo, err)
	}
}

func (s *Service) NotifyAdmins(ctx context.Context, input NotificationInput) {
	adminIDs, err := s.repository.GetAdminUserIDs(ctx)
	if err != nil {
		log.Printf("notifications: no se pudieron obtener los administradores: %v", err)
		return
	}

	for _, adminID := range adminIDs {
		adminInput := input
		adminInput.UserID = adminID
		s.Notify(ctx, adminInput)
	}
}

// GetNotifications retorna una página del buzón del usuario junto con el total
// de notificaciones que posee, para que el cliente sepa si hay más.
func (s *Service) GetNotifications(ctx context.Context, userID uint, limit, offset int) ([]Notification, int, error) {
	if userID == 0 {
		return nil, 0, ErrUserRequired
	}
	limit, offset = NormalizePagination(limit, offset)

	items, err := s.repository.GetByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repository.CountByUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID uint) (int, error) {
	if userID == 0 {
		return 0, ErrUserRequired
	}
	return s.repository.GetUnreadCount(ctx, userID)
}

func (s *Service) MarkAllRead(ctx context.Context, userID uint) (int, error) {
	if userID == 0 {
		return 0, ErrUserRequired
	}
	return s.repository.MarkAllRead(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, id int, userID uint) error {
	if userID == 0 {
		return ErrUserRequired
	}
	if id <= 0 {
		return errors.New("ID de notificación inválido")
	}
	return s.repository.MarkRead(ctx, id, userID)
}
