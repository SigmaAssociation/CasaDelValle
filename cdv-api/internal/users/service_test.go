package users

import (
	"context"
	"testing"

	"cdv-api/internal/middleware"
	"cdv-api/internal/notifications"
)

type mockUserRepository struct {
	createReturnID  int
	createReturnErr error
	updateRoleRows  int
	updateRoleErr   error
	getByIDReturn   User
	getByIDErr      error
}

func (m *mockUserRepository) CreateUser(ctx context.Context, req RegisterRequest, hashPassword string) (int, error) {
	return m.createReturnID, m.createReturnErr
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, userUpdate UserUpdate) (int, error) {
	return 0, nil
}

func (m *mockUserRepository) UpdateRole(ctx context.Context, userID int, role uint) (int, error) {
	return m.updateRoleRows, m.updateRoleErr
}

func (m *mockUserRepository) GetAll(ctx context.Context) ([]User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetUserAuthInfo(ctx context.Context, email string) (UserAuth, error) {
	return UserAuth{}, nil
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id int) (User, error) {
	return m.getByIDReturn, m.getByIDErr
}

type mockUserNotifier struct {
	inputs      []notifications.NotificationInput
	adminInputs []notifications.NotificationInput
}

func (m *mockUserNotifier) Notify(ctx context.Context, input notifications.NotificationInput) {
	m.inputs = append(m.inputs, input)
}

func (m *mockUserNotifier) NotifyAdmins(ctx context.Context, input notifications.NotificationInput) {
	m.adminInputs = append(m.adminInputs, input)
}

func (m *mockUserNotifier) byTipo(tipo string) []notifications.NotificationInput {
	var out []notifications.NotificationInput
	for _, in := range m.inputs {
		if in.Tipo == tipo {
			out = append(out, in)
		}
	}
	return out
}

func validRegisterRequest() RegisterRequest {
	return RegisterRequest{
		Name:     "Juan Perez",
		DPI:      "1234567890123",
		Phone:    "12345678",
		Email:    "juan@example.com",
		Address:  "5a avenida 10-20 zona 1",
		Password: "Password1!",
		IDRole:   middleware.RoleGuest,
	}
}

func TestRegisterUserNotifiesWelcomeAndAdmins(t *testing.T) {
	repo := &mockUserRepository{createReturnID: 10}
	notifier := &mockUserNotifier{}
	service := NewService(repo, notifier)

	id, err := service.RegisterUser(context.Background(), validRegisterRequest())
	if err != nil {
		t.Fatalf("RegisterUser returned error: %v", err)
	}
	if id != 10 {
		t.Fatalf("RegisterUser() = %d, want 10", id)
	}

	welcome := notifier.byTipo(notifications.TipoBienvenida)
	if len(welcome) != 1 || welcome[0].UserID != 10 {
		t.Fatalf("expected 1 welcome notification for user 10, got %+v", welcome)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}

func TestUpgradeToHostNotifiesUserAndAdmins(t *testing.T) {
	repo := &mockUserRepository{
		getByIDReturn:  User{ID: 5, Name: "Ana", IDRole: middleware.RoleGuest},
		updateRoleRows: 1,
	}
	notifier := &mockUserNotifier{}
	service := NewService(repo, notifier)

	auth, err := service.UpgradeToHost(context.Background(), 5)
	if err != nil {
		t.Fatalf("UpgradeToHost returned error: %v", err)
	}
	if auth.IDRole != middleware.RoleHost {
		t.Fatalf("UpgradeToHost() role = %d, want %d", auth.IDRole, middleware.RoleHost)
	}

	role := notifier.byTipo(notifications.TipoRolActualizado)
	if len(role) != 1 || role[0].UserID != 5 {
		t.Fatalf("expected 1 role notification for user 5, got %+v", role)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}
