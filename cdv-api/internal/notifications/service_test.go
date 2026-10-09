package notifications

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockNotificationRepository struct {
	createCalled    bool
	createInput     NotificationInput
	createInputs    []NotificationInput
	createReturnID  int
	createReturnErr error

	getByUserReturn  []Notification
	getByUserErr     error
	getByUserLimit   int
	getByUserOffset  int
	countByUserTotal int
	countByUserErr   error

	adminIDsReturn []uint
	adminIDsErr    error

	unreadCountReturn int
	unreadCountErr    error

	markAllReturn int
	markAllErr    error

	markReadErr error
}

func (m *mockNotificationRepository) Create(ctx context.Context, input NotificationInput) (int, error) {
	m.createCalled = true
	m.createInput = input
	m.createInputs = append(m.createInputs, input)
	return m.createReturnID, m.createReturnErr
}

func (m *mockNotificationRepository) GetByUser(ctx context.Context, userID uint, limit, offset int) ([]Notification, error) {
	m.getByUserLimit = limit
	m.getByUserOffset = offset
	return m.getByUserReturn, m.getByUserErr
}

func (m *mockNotificationRepository) CountByUser(ctx context.Context, userID uint) (int, error) {
	return m.countByUserTotal, m.countByUserErr
}

func (m *mockNotificationRepository) GetUnreadCount(ctx context.Context, userID uint) (int, error) {
	return m.unreadCountReturn, m.unreadCountErr
}

func (m *mockNotificationRepository) GetAdminUserIDs(ctx context.Context) ([]uint, error) {
	return m.adminIDsReturn, m.adminIDsErr
}

func (m *mockNotificationRepository) MarkAllRead(ctx context.Context, userID uint) (int, error) {
	return m.markAllReturn, m.markAllErr
}

func (m *mockNotificationRepository) MarkRead(ctx context.Context, id int, userID uint) error {
	return m.markReadErr
}

func validInput() NotificationInput {
	return NotificationInput{
		UserID:      1,
		Tipo:        TipoReservaCreada,
		Mensaje:     "Tu reserva fue registrada exitosamente",
		EntidadTipo: EntidadReservacion,
		EntidadID:   10,
	}
}

func TestCreateRejectsMissingUser(t *testing.T) {
	repo := &mockNotificationRepository{}
	service := NewService(repo)

	_, err := service.Create(context.Background(), NotificationInput{
		UserID:  0,
		Tipo:    TipoBienvenida,
		Mensaje: "Bienvenido",
	})
	if err == nil {
		t.Fatal("expected error for missing user")
	}
	if repo.createCalled {
		t.Fatal("repository.Create should not be called when user is missing")
	}
}

func TestCreateRejectsUnknownTipo(t *testing.T) {
	repo := &mockNotificationRepository{}
	service := NewService(repo)

	input := validInput()
	input.Tipo = "tipo_inexistente"

	_, err := service.Create(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for unknown tipo")
	}
	if !strings.Contains(err.Error(), "no válido") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalled {
		t.Fatal("repository.Create should not be called for an unknown tipo")
	}
}

func TestCreateRejectsEmptyMessage(t *testing.T) {
	repo := &mockNotificationRepository{}
	service := NewService(repo)

	input := validInput()
	input.Mensaje = "   "

	_, err := service.Create(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty message")
	}
	if repo.createCalled {
		t.Fatal("repository.Create should not be called for an empty message")
	}
}

func TestCreateRejectsUnknownEntidad(t *testing.T) {
	repo := &mockNotificationRepository{}
	service := NewService(repo)

	input := validInput()
	input.EntidadTipo = "inexistente"

	_, err := service.Create(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for unknown entidad tipo")
	}
	if repo.createCalled {
		t.Fatal("repository.Create should not be called for an unknown entidad tipo")
	}
}

func TestCreateAllowsEmptyEntidad(t *testing.T) {
	repo := &mockNotificationRepository{createReturnID: 7}
	service := NewService(repo)

	input := validInput()
	input.EntidadTipo = ""
	input.EntidadID = 0

	id, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id != 7 {
		t.Fatalf("Create() = %d, want 7", id)
	}
	if !repo.createCalled {
		t.Fatal("repository.Create should be called when entidad is empty")
	}
}

func TestNotifyDoesNotPanicOnInvalidInput(t *testing.T) {
	repo := &mockNotificationRepository{}
	service := NewService(repo)
	service.Notify(context.Background(), NotificationInput{UserID: 0})
	if repo.createCalled {
		t.Fatal("repository.Create should not be called for invalid input")
	}
}

func TestNotifyAdminsReplicatesForEachAdmin(t *testing.T) {
	repo := &mockNotificationRepository{adminIDsReturn: []uint{1, 4, 9}, createReturnID: 1}
	service := NewService(repo)

	input := validInput()
	input.UserID = 0
	service.NotifyAdmins(context.Background(), input)
	if len(repo.createInputs) != 3 {
		t.Fatalf("expected 3 admin notifications, got %d", len(repo.createInputs))
	}
	for i, adminID := range []uint{1, 4, 9} {
		if repo.createInputs[i].UserID != adminID {
			t.Fatalf("admin notification %d should target user %d, got %d", i, adminID, repo.createInputs[i].UserID)
		}
	}
}

func TestNotifyAdminsIgnoresLookupError(t *testing.T) {
	repo := &mockNotificationRepository{adminIDsErr: errors.New("boom")}
	service := NewService(repo)
	service.NotifyAdmins(context.Background(), validInput())

	if repo.createCalled {
		t.Fatal("repository.Create should not be called when admin lookup fails")
	}
}

func TestGetNotificationsRejectsMissingUser(t *testing.T) {
	service := NewService(&mockNotificationRepository{})

	_, _, err := service.GetNotifications(context.Background(), 0, 0, 0)
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

func TestGetNotificationsReturnsListAndTotal(t *testing.T) {
	expected := []Notification{
		{ID: 2, UserID: 1, Tipo: TipoReservaCreada, Mensaje: "más reciente"},
		{ID: 1, UserID: 1, Tipo: TipoBienvenida, Mensaje: "más antigua"},
	}
	repo := &mockNotificationRepository{getByUserReturn: expected, countByUserTotal: 7}
	service := NewService(repo)

	items, total, err := service.GetNotifications(context.Background(), 1, 2, 0)
	if err != nil {
		t.Fatalf("GetNotifications returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("GetNotifications() returned %d items, want 2", len(items))
	}
	if items[0].ID != 2 {
		t.Fatalf("GetNotifications() should keep repository order, got id %d first", items[0].ID)
	}
	if total != 7 {
		t.Fatalf("GetNotifications() total = %d, want 7", total)
	}
}

func TestGetNotificationsNormalizesPagination(t *testing.T) {
	repo := &mockNotificationRepository{}
	service := NewService(repo)

	// límite negativo y offset negativo se normalizan.
	if _, _, err := service.GetNotifications(context.Background(), 1, -5, -10); err != nil {
		t.Fatalf("GetNotifications returned error: %v", err)
	}
	if repo.getByUserLimit != DefaultPageLimit {
		t.Fatalf("limit = %d, want %d", repo.getByUserLimit, DefaultPageLimit)
	}
	if repo.getByUserOffset != 0 {
		t.Fatalf("offset = %d, want 0", repo.getByUserOffset)
	}

	// un límite mayor al máximo se recorta.
	if _, _, err := service.GetNotifications(context.Background(), 1, 9999, 5); err != nil {
		t.Fatalf("GetNotifications returned error: %v", err)
	}
	if repo.getByUserLimit != MaxPageLimit {
		t.Fatalf("limit = %d, want %d", repo.getByUserLimit, MaxPageLimit)
	}
	if repo.getByUserOffset != 5 {
		t.Fatalf("offset = %d, want 5", repo.getByUserOffset)
	}
}

func TestNormalizePagination(t *testing.T) {
	cases := []struct {
		limit, offset         int
		wantLimit, wantOffset int
	}{
		{0, 0, DefaultPageLimit, 0},
		{-1, -1, DefaultPageLimit, 0},
		{10, 20, 10, 20},
		{1000, 0, MaxPageLimit, 0},
	}
	for _, tc := range cases {
		gotLimit, gotOffset := NormalizePagination(tc.limit, tc.offset)
		if gotLimit != tc.wantLimit || gotOffset != tc.wantOffset {
			t.Fatalf("NormalizePagination(%d, %d) = (%d, %d), want (%d, %d)",
				tc.limit, tc.offset, gotLimit, gotOffset, tc.wantLimit, tc.wantOffset)
		}
	}
}

func TestGetUnreadCountRejectsMissingUser(t *testing.T) {
	service := NewService(&mockNotificationRepository{})

	_, err := service.GetUnreadCount(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

func TestMarkAllReadReturnsUpdatedCount(t *testing.T) {
	repo := &mockNotificationRepository{markAllReturn: 3}
	service := NewService(repo)

	updated, err := service.MarkAllRead(context.Background(), 5)
	if err != nil {
		t.Fatalf("MarkAllRead returned error: %v", err)
	}
	if updated != 3 {
		t.Fatalf("MarkAllRead() = %d, want 3", updated)
	}
}

func TestMarkReadRejectsInvalidID(t *testing.T) {
	repo := &mockNotificationRepository{}
	service := NewService(repo)

	err := service.MarkRead(context.Background(), 0, 1)
	if err == nil {
		t.Fatal("expected error for invalid id")
	}
}

func TestMarkReadNotFound(t *testing.T) {
	repo := &mockNotificationRepository{markReadErr: ErrNotificationNotFound}
	service := NewService(repo)

	err := service.MarkRead(context.Background(), 99, 1)
	if err == nil {
		t.Fatal("expected ErrNotificationNotFound")
	}
}
