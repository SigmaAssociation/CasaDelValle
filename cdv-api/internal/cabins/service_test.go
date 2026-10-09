package cabins

import (
	"context"
	"testing"

	"cdv-api/internal/notifications"
)

type mockCabinRepository struct {
	createReturnID  int
	createReturnErr error
	getByIDReturn   Cabin
	getByIDErr      error
	updateErr       error
	deleteReturn    int
	deleteErr       error
}

func (m *mockCabinRepository) GetAll(ctx context.Context) ([]Cabin, error) {
	return nil, nil
}

func (m *mockCabinRepository) CreateCabin(ctx context.Context, req CreateCabinRequest) (int, error) {
	return m.createReturnID, m.createReturnErr
}

func (m *mockCabinRepository) GetByID(ctx context.Context, id int) (Cabin, error) {
	return m.getByIDReturn, m.getByIDErr
}

func (m *mockCabinRepository) Update(ctx context.Context, id int, req UpdateCabinRequest) error {
	return m.updateErr
}

func (m *mockCabinRepository) DeleteCabin(ctx context.Context, id int) (int, error) {
	return m.deleteReturn, m.deleteErr
}

func (m *mockCabinRepository) GetByHostID(ctx context.Context, hostID int) ([]Cabin, error) {
	return nil, nil
}

func (m *mockCabinRepository) SearchCabins(ctx context.Context, params CabinSearchParams) ([]CabinCardResponse, error) {
	return nil, nil
}

type mockCabinNotifier struct {
	inputs      []notifications.NotificationInput
	adminInputs []notifications.NotificationInput
}

func (m *mockCabinNotifier) Notify(ctx context.Context, input notifications.NotificationInput) {
	m.inputs = append(m.inputs, input)
}

func (m *mockCabinNotifier) NotifyAdmins(ctx context.Context, input notifications.NotificationInput) {
	m.adminInputs = append(m.adminInputs, input)
}

func (m *mockCabinNotifier) byTipo(tipo string) []notifications.NotificationInput {
	var out []notifications.NotificationInput
	for _, in := range m.inputs {
		if in.Tipo == tipo {
			out = append(out, in)
		}
	}
	return out
}

func validCabinRequest() CreateCabinRequest {
	return CreateCabinRequest{
		Name:        "Cabaña Lago",
		Address:     "Kilómetro 12, carretera al lago",
		Price:       150,
		Description: "Vista al lago",
		Capacity:    4,
		Rules:       "No fumar",
		HostID:      4,
	}
}

func TestCreateCabinNotifiesHostAndAdmins(t *testing.T) {
	repo := &mockCabinRepository{createReturnID: 21}
	notifier := &mockCabinNotifier{}
	service := NewService(repo, notifier)

	id, err := service.CreateCabin(context.Background(), validCabinRequest())
	if err != nil {
		t.Fatalf("CreateCabin returned error: %v", err)
	}
	if id != 21 {
		t.Fatalf("CreateCabin() = %d, want 21", id)
	}

	created := notifier.byTipo(notifications.TipoCabanaCreada)
	if len(created) != 1 || created[0].UserID != 4 {
		t.Fatalf("expected 1 notification for host 4, got %+v", created)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}

func TestUpdateCabinNotifiesHostAndAdmins(t *testing.T) {
	repo := &mockCabinRepository{getByIDReturn: Cabin{ID: 7, Name: "Cabaña Sol", HostID: 4}}
	notifier := &mockCabinNotifier{}
	service := NewService(repo, notifier)

	req := UpdateCabinRequest{
		Name:        "Cabaña Sol",
		Address:     "Kilómetro 12, carretera al lago",
		Price:       175,
		Description: "Remodelada",
		Capacity:    5,
		Rules:       "No fumar",
	}
	if err := service.UpdateCabin(context.Background(), 7, req); err != nil {
		t.Fatalf("UpdateCabin returned error: %v", err)
	}

	updated := notifier.byTipo(notifications.TipoCabanaActualizada)
	if len(updated) != 1 || updated[0].UserID != 4 {
		t.Fatalf("expected 1 notification for host 4, got %+v", updated)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}

func TestDeleteCabinNotifiesHostAndAdmins(t *testing.T) {
	repo := &mockCabinRepository{
		getByIDReturn: Cabin{ID: 9, Name: "Cabaña Mar", HostID: 4},
		deleteReturn:  1,
	}
	notifier := &mockCabinNotifier{}
	service := NewService(repo, notifier)

	rows, err := service.DeleteCabin(context.Background(), 9)
	if err != nil {
		t.Fatalf("DeleteCabin returned error: %v", err)
	}
	if rows != 1 {
		t.Fatalf("DeleteCabin() = %d, want 1", rows)
	}

	deleted := notifier.byTipo(notifications.TipoCabanaEliminada)
	if len(deleted) != 1 || deleted[0].UserID != 4 {
		t.Fatalf("expected 1 notification for host 4, got %+v", deleted)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}
