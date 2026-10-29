package payments

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"cdv-api/internal/middleware"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes mantiene juntas la consulta y el registro de abonos.
func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /cdv-api/payments/reservation/{reservationId}", c.GetReservationPayments)
	mux.HandleFunc("POST /cdv-api/payments/reservation/{reservationId}", c.CreatePayment)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}

// CreatePayment registra un pago/abono de una reservación.
// @Summary Registrar pago de una reservación
// @Description Registra un pago simulado (tarjeta, transferencia o efectivo) sobre una reservación confirmada o en curso. El monto debe ser mayor que 0 y no puede exceder el saldo pendiente. Puede registrarlo quien reservó, el dueño de la cabaña o un admin. Requiere token.
// @Tags Pagos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reservationId path int true "ID de reservación"
// @Param payment body CreatePaymentRequest true "Monto y método de pago"
// @Success 201 {object} CreatePaymentResponse "Pago registrado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos, monto o método no válidos"
// @Failure 401 {object} map[string]string "No autenticado"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Reservación no encontrada"
// @Failure 409 {object} map[string]string "La reservación no admite pagos o el monto excede el saldo"
// @Router /payments/reservation/{reservationId} [post]
func (c *Controller) CreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id, ok := parseReservationID(w, r)
	if !ok {
		return
	}

	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}

	info, err := c.service.GetReservationInfo(r.Context(), id)
	if err != nil {
		writeLookupError(w, err)
		return
	}

	if !canAccessPayments(r, info) {
		writeError(w, http.StatusForbidden, "No tiene permisos para registrar pagos en esta reservación")
		return
	}

	result, err := c.service.CreatePayment(r.Context(), id, actorID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAmount):
			writeError(w, http.StatusBadRequest, "El monto debe ser mayor que 0")
		case errors.Is(err, ErrInvalidMethod):
			writeError(w, http.StatusBadRequest, "Método de pago inválido")
		case errors.Is(err, ErrAmountExceedsBalance):
			writeError(w, http.StatusConflict, "El monto no puede superar el saldo pendiente")
		case errors.Is(err, ErrReservationNotPayable):
			writeError(w, http.StatusConflict, "La reservación no admite pagos en su estado actual")
		case errors.Is(err, ErrReservationNotFound):
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
		default:
			writeError(w, http.StatusInternalServerError, "Error al registrar el pago")
		}
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetReservationPayments lista los pagos de una reservación.
// @Summary Listar pagos de una reservación
// @Description Devuelve el historial de pagos y el desglose de total, pagado y saldo pendiente. Puede verlo quien reservó, el dueño de la cabaña o un admin. Requiere token.
// @Tags Pagos
// @Produce json
// @Security BearerAuth
// @Param reservationId path int true "ID de reservación"
// @Success 200 {object} ReservationPayments "Historial de pagos y saldo"
// @Failure 400 {object} map[string]string "ID de reservación inválido"
// @Failure 401 {object} map[string]string "No autenticado"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Reservación no encontrada"
// @Router /payments/reservation/{reservationId} [get]
func (c *Controller) GetReservationPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id, ok := parseReservationID(w, r)
	if !ok {
		return
	}

	info, err := c.service.GetReservationInfo(r.Context(), id)
	if err != nil {
		writeLookupError(w, err)
		return
	}

	if !canAccessPayments(r, info) {
		writeError(w, http.StatusForbidden, "No tiene permisos para ver los pagos de esta reservación")
		return
	}

	summary, err := c.service.GetReservationPayments(r.Context(), id)
	if err != nil {
		writeLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func parseReservationID(w http.ResponseWriter, r *http.Request) (int, bool) {
	idParam := r.PathValue("reservationId")
	if idParam == "" {
		idParam = r.URL.Query().Get("reservationId")
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de reservación inválido")
		return 0, false
	}
	return id, true
}

func writeLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrReservationNotFound) {
		writeError(w, http.StatusNotFound, "Reservación no encontrada")
		return
	}
	writeError(w, http.StatusInternalServerError, "Error al obtener la reservación")
}

// canAccessPayments indica si el usuario autenticado puede ver o registrar
// pagos: quien reservó, el dueño de la cabaña o un administrador.
func canAccessPayments(r *http.Request, info *ReservationPaymentInfo) bool {
	return middleware.AuthorizeSelfOrAdmin(r, info.UserID) ||
		middleware.AuthorizeSelfOrAdmin(r, info.HostID)
}
