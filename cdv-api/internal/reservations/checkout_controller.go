package reservations

import (
	"errors"
	"net/http"
	"strconv"

	"cdv-api/internal/middleware"
)

// Checkout registra el check-out de una reservación.
// @Summary Registrar check-out
// @Description Finaliza la estadía de una reservación en curso que ya tiene check-in registrado y la pasa a finalizada. Solo el anfitrión de la cabaña o un admin. Requiere token.
// @Tags Reservaciones
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de reservación"
// @Success 200 {object} CheckoutReservationResponse "Check-out registrado exitosamente"
// @Failure 400 {object} map[string]string "ID de reservación inválido"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Reservación no encontrada"
// @Failure 409 {object} map[string]string "Sin check-in o reservación no en curso"
// @Router /reservations/{id}/checkout [patch]
func (c *Controller) Checkout(w http.ResponseWriter, r *http.Request) {
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

	// Solo el anfitrión de la cabaña (o un admin) registra el check-out.
	if !middleware.AuthorizeSelfOrAdmin(r, reservation.HostID) {
		writeError(w, http.StatusForbidden, "No tiene permisos para registrar el check-out de esta reservación")
		return
	}

	finished, err := c.service.Checkout(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrReservationNotFound):
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
		case errors.Is(err, ErrCheckinRequired):
			writeError(w, http.StatusConflict, "No se puede registrar el check-out sin un check-in previo")
		case errors.Is(err, ErrReservationNotCheckoutable):
			writeError(w, http.StatusConflict, "Solo se puede registrar el check-out de una reservación en curso")
		default:
			writeError(w, http.StatusInternalServerError, "Error al registrar el check-out")
		}
		return
	}

	writeJSON(w, http.StatusOK, CheckoutReservationResponse{
		Message:     "Check-out registrado exitosamente",
		Reservation: finished,
	})
}
