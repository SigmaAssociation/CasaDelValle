package reservations

import (
	"errors"
	"net/http"
	"strconv"

	"cdv-api/internal/middleware"
)

// GetReservationBalance retorna el estado financiero de una reservación.
// @Summary Consultar saldo de una reservación
// @Description Retorna total, cargos adicionales, pagos, reembolsos y saldo pendiente. Puede verlo quien reservó, el anfitrión de la cabaña o un admin. Requiere token.
// @Tags Reservaciones
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de reservación"
// @Success 200 {object} ReservationBalance "Saldo de la reservación"
// @Failure 400 {object} map[string]string "ID de reservación inválido"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Reservación no encontrada"
// @Router /reservations/{id}/balance [get]
func (c *Controller) GetReservationBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	canView := middleware.AuthorizeSelfOrAdmin(r, reservation.UserID) ||
		middleware.AuthorizeSelfOrAdmin(r, reservation.HostID)
	if !canView {
		writeError(w, http.StatusForbidden, "No tiene permisos para ver el saldo de esta reservación")
		return
	}

	balance, err := c.service.GetReservationBalance(r.Context(), reservation)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) || errors.Is(err, ErrCabinNotFound) {
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al calcular el saldo de la reservación")
		return
	}

	writeJSON(w, http.StatusOK, balance)
}
