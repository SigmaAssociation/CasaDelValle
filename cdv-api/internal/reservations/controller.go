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

// CreateReservation registra una nueva reservación.
// @Summary Crear reservación
// @Description Crea una reservación. Para clientes el usuario se toma del token; para administradores puede indicarse. Requiere token.
// @Tags Reservaciones
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reservation body ReservationRequest true "Datos de la reservación"
// @Success 201 {object} CreateReservationResponse "Reservación registrada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Cabaña o usuario no existe"
// @Failure 409 {object} map[string]string "Conflicto de fechas"
// @Router /reservations [post]
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

// CancelReservation cancela una reservación.
// @Summary Cancelar reservación
// @Description Cancela una reservación activa. Puede cancelar quien reservó, el dueño de la cabaña o un admin, siempre que falten al menos 3 días para el inicio. Requiere token.
// @Tags Reservaciones
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de reservación"
// @Success 200 {object} CancelReservationResponse "Reservación cancelada exitosamente"
// @Failure 400 {object} map[string]string "ID de reservación inválido"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Reservación no encontrada"
// @Failure 409 {object} map[string]string "No cancelable o con menos de 3 días de anticipación"
// @Router /reservations/{id}/cancel [patch]
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

// GetReservationByID retorna el detalle de una reservación.
// @Summary Obtener reservación por ID
// @Description Retorna el detalle de una reservación. Puede verla quien reservó, el dueño de la cabaña o un admin. Requiere token.
// @Tags Reservaciones
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de reservación"
// @Success 200 {object} ReservationDetail "Detalle de la reservación"
// @Failure 400 {object} map[string]string "ID de reservación inválido"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Reservación no encontrada"
// @Router /reservations/{id} [get]
func (c *Controller) GetReservationByID(w http.ResponseWriter, r *http.Request) {
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

	// Puede verla: quien reservó, el dueño de la cabaña o un administrador.
	canView := middleware.AuthorizeSelfOrAdmin(r, reservation.UserID) ||
		middleware.AuthorizeSelfOrAdmin(r, reservation.HostID)
	if !canView {
		writeError(w, http.StatusForbidden, "No tiene permisos para ver esta reservación")
		return
	}

	writeJSON(w, http.StatusOK, reservation)
}

func (c *Controller) GetReservationsByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("userId")
	userID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || userID == 0 {
		writeError(w, http.StatusBadRequest, "ID de usuario inválido")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(userID)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para ver estas reservaciones")
		return
	}

	cards, err := c.service.GetReservationsByUserID(r.Context(), uint(userID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener las reservaciones")
		return
	}

	writeJSON(w, http.StatusOK, cards)
}

// GetReservationsByCabin retorna las reservaciones de una cabaña.
// @Summary Obtener reservaciones por cabaña
// @Description Retorna el listado de reservaciones de una cabaña. Puede verla quien reservó, el dueño de la cabaña o un admin. Requiere token.
// @Tags Reservaciones, Cabañas
// @Produce json
// @Security BearerAuth
// @Param cabinId path int true "ID de cabaña"
// @Success 200 {object} ReservationCard[] "Listado de reservaciones de la cabaña"
// @Router /reservations/cabin/{cabinId} [get]
func (c *Controller) GetReservationsByCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("cabinId")
	cabinID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || cabinID == 0 {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	hostID, err := c.service.GetCabinHostID(r.Context(), uint(cabinID))
	if err != nil {
		if errors.Is(err, ErrCabinNotFound) {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener las reservaciones")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, hostID) {
		writeError(w, http.StatusForbidden, "No tiene permisos para ver estas reservaciones")
		return
	}

	cards, err := c.service.GetReservationsByCabinID(r.Context(), uint(cabinID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener las reservaciones")
		return
	}

	writeJSON(w, http.StatusOK, cards)
}

// UpdateReservation actualiza una reservación.
// @Summary Actualizar reservación
// @Description Actualiza la cabaña y las fechas de una reservación activa. Puede editar quien reservó, el dueño de la cabaña o un admin. Requiere token.
// @Tags Reservaciones
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de reservación"
// @Param reservation body UpdateReservationRequest true "Datos a actualizar"
// @Success 200 {object} UpdateReservationResponse "Reservación actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Reservación no encontrada"
// @Failure 409 {object} map[string]string "Conflicto de fechas o reservación no editable"
// @Router /reservations/{id} [put]
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
