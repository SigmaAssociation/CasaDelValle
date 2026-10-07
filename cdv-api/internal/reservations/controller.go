package reservations

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
		case "La cabaña no está disponible en el rango de fechas seleccionado",
			"El usuario ya tiene otra reservación en el rango de fechas seleccionado":
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

func (c *Controller) CancelReservation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		idParam = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de reservación inválido")
		return
	}

	reservation, err := c.service.GetReservationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) {
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la reservación")
		return
	}

	// Pueden cancelar: quien reservó, el dueño de la cabaña o un administrador.
	canCancel := middleware.AuthorizeSelfOrAdmin(r, reservation.UserID) ||
		middleware.AuthorizeSelfOrAdmin(r, reservation.HostID)
	if !canCancel {
		writeError(w, http.StatusForbidden, "No tiene permisos para cancelar esta reservación")
		return
	}

	cancelled, err := c.service.CancelReservation(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrCancellationTooLate):
			writeError(w, http.StatusConflict, "Solo se puede cancelar con al menos 3 días de anticipación")
		case errors.Is(err, ErrReservationNotCancellable):
			writeError(w, http.StatusConflict, "La reservación no se puede cancelar")
		case errors.Is(err, ErrReservationNotFound):
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
		default:
			writeError(w, http.StatusInternalServerError, "Error al cancelar la reservación")
		}
		return
	}

	writeJSON(w, http.StatusOK, CancelReservationResponse{
		Message:     "Reservación cancelada exitosamente",
		Reservation: cancelled,
	})
}

func (c *Controller) UpdateReservation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		idParam = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de reservación inválido")
		return
	}

	var req UpdateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}

	existing, err := c.service.GetReservationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) {
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la reservación")
		return
	}

	// Pueden editar: quien reservó, el dueño de la cabaña o un administrador.
	canEdit := middleware.AuthorizeSelfOrAdmin(r, existing.UserID) ||
		middleware.AuthorizeSelfOrAdmin(r, existing.HostID)
	if !canEdit {
		writeError(w, http.StatusForbidden, "No tiene permisos para editar esta reservación")
		return
	}

	updated, err := c.service.UpdateReservation(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrReservationNotFound):
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
		case errors.Is(err, ErrReservationNotEditable):
			writeError(w, http.StatusConflict, "La reservación no se puede modificar")
		case err.Error() == "La cabaña no está disponible en el rango de fechas seleccionado" ||
			err.Error() == "El usuario ya tiene otra reservación en el rango de fechas seleccionado":
			writeError(w, http.StatusConflict, err.Error())
		case err.Error() == "La cabaña especificada no existe":
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, UpdateReservationResponse{
		Message:     "Reservación actualizada exitosamente",
		Reservation: updated,
	})
}
