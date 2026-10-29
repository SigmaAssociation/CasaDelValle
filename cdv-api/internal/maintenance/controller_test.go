package maintenance

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cdv-api/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func TestControllerAllowsHostAndAdminRoles(t *testing.T) {
	t.Setenv("JWT_SECRET", "maintenance-test")
	for _, tt := range []struct {
		role int
		code int
	}{{0, 401}, {1, 200}, {2, 403}, {3, 200}} {
		repo := &stubRepository{}
		controller := NewController(NewService(repo))
		mux := http.NewServeMux()
		mux.HandleFunc("POST /cdv-api/maintenance/cabin/{cabinId}/preview", controller.Preview)
		req := httptest.NewRequest("POST", "/cdv-api/maintenance/cabin/100/preview", strings.NewReader(`{"start_date":"2099-01-01","end_date":"2099-01-03"}`))
		if tt.role != 0 {
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": 10, "role": tt.role}).SignedString([]byte("maintenance-test"))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		middleware.Auth(mux).ServeHTTP(response, req)
		if response.Code != tt.code {
			t.Fatalf("role %d: %d %s", tt.role, response.Code, response.Body)
		}
		if repo.called != (tt.role == 3 || tt.role == 1) {
			t.Fatalf("acceso a repositorio sin permisos: %d", tt.role)
		}
	}
}
func TestUnexpectedErrorsStayPrivate(t *testing.T) {
	response := httptest.NewRecorder()
	fail(response, errors.New("SQL secret internal detail"))
	if response.Code != 500 || strings.Contains(response.Body.String(), "secret") {
		t.Fatal(response.Code, response.Body)
	}
}
