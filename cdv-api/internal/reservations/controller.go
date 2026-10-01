package reservations

import (
	"encoding/json"
	"net/http"

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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Message: message})
}

func (c *Controller) CreateReservation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var req ReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}

	userIDUint, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	if !middleware.IsAdmin(r) {
		req.UserID = userIDUint
	} else if req.UserID == 0 {
		req.UserID = userIDUint
	}

	if !middleware.AuthorizeSelfOrAdmin(r, req.UserID) {
		writeError(w, http.StatusForbidden, "No tiene permisos para registrar esta reservación")
		return
	}

	reservationID, err := c.service.CreateReservation(r.Context(), req)
	if err != nil {
		switch err.Error() {
		case "La cabaña no está disponible en el rango de fechas seleccionado":
			writeError(w, http.StatusConflict, err.Error())
		case "La cabaña especificada no existe", "El usuario especificado no existe":
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, CreateReservationResponse{
		Message:       "Reservación registrada exitosamente",
		ReservationID: reservationID,
	})
}
