package payments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cdv-api/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func TestRegisteredRoutesAcceptPaymentsAfterAnticipo(t *testing.T) {
	t.Setenv("JWT_SECRET", "payment-route-test")
	for _, method := range []string{MethodCard, MethodTransfer, MethodCash} {
		t.Run(method, func(t *testing.T) {
			info := confirmedReservation()
			contractTotal := 808.33
			info.TotalPrice = &contractTotal
			repo := &mockPaymentRepository{info: info, paid: 242.50}
			mux := http.NewServeMux()
			NewController(NewService(repo, nil)).RegisterRoutes(mux)
			handler := middleware.Auth(mux)
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": 5, "role": 2}).SignedString([]byte("payment-route-test"))
			if err != nil {
				t.Fatal(err)
			}
			request := func(amount float64) *httptest.ResponseRecorder {
				req := httptest.NewRequest("POST", "/cdv-api/payments/reservation/10", strings.NewReader(fmt.Sprintf(`{"amount":%.2f,"method":%q}`, amount, method)))
				req.Header.Set("Authorization", "Bearer "+token)
				res := httptest.NewRecorder()
				handler.ServeHTTP(res, req)
				return res
			}
			first := request(100)
			if first.Code != http.StatusCreated {
				t.Fatalf("primer abono: %d %s", first.Code, first.Body)
			}
			var result CreatePaymentResponse
			if err := json.Unmarshal(first.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Total != 808.33 || result.Paid != 342.50 || result.Balance != 465.83 {
				t.Fatalf("resumen incorrecto: %+v", result)
			}
			final := request(465.83)
			if final.Code != http.StatusCreated {
				t.Fatalf("pago de saldo: %d %s", final.Code, final.Body)
			}
			if err := json.Unmarshal(final.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Balance != 0 || result.Paid != 808.33 || len(repo.calls) != 2 {
				t.Fatalf("saldo no completado: %+v", result)
			}
			if extra := request(1); extra.Code != http.StatusConflict {
				t.Fatalf("permitió exceder saldo: %d", extra.Code)
			}
		})
	}
}
