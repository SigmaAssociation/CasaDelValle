package maintenance

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"cdv-api/internal/middleware"
)

type Controller struct{ service *Service }

func NewController(s *Service) *Controller { return &Controller{service: s} }
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	var invalid validationError
	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, ErrConflict), errors.Is(err, ErrUnavailable):
		status = http.StatusConflict
	case errors.As(err, &invalid):
	default:
		log.Printf("maintenance: %v", err)
		respond(w, http.StatusInternalServerError, ErrorResponse{"No se pudo completar la operación de mantenimiento"})
		return
	}
	respond(w, status, ErrorResponse{err.Error()})
}
func actor(w http.ResponseWriter, r *http.Request) (int, uint, bool) {
	user, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond(w, 401, ErrorResponse{"Usuario no autenticado"})
		return 0, 0, false
	}
	if !middleware.IsHostOrAdmin(r) {
		respond(w, 403, ErrorResponse{"Solo un anfitrión o administrador puede gestionar mantenimientos"})
		return 0, 0, false
	}
	id, err := strconv.Atoi(r.PathValue("cabinId"))
	if err != nil || id <= 0 {
		respond(w, 400, ErrorResponse{"ID de cabaña inválido"})
		return 0, 0, false
	}
	return id, user, true
}

// List godoc
// @Summary Listar los mantenimientos de una cabaña propia
// @Tags Mantenimientos
// @Produce json
// @Security BearerAuth
// @Param cabinId path int true "ID de cabaña"
// @Success 200 {array} Maintenance
// @Failure 403 {object} ErrorResponse
// @Router /maintenance/cabin/{cabinId} [get]
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	id, user, ok := actor(w, r)
	if !ok {
		return
	}
	result, err := c.service.List(r.Context(), id, user)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, result)
}

// Preview godoc
// @Summary Consultar las reservas que se notificarán por mantenimiento
// @Tags Mantenimientos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cabinId path int true "ID de cabaña"
// @Param body body DateRequest true "Rango de mantenimiento [inicio, fin)"
// @Success 200 {object} Preview
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /maintenance/cabin/{cabinId}/preview [post]
func (c *Controller) Preview(w http.ResponseWriter, r *http.Request) {
	id, user, ok := actor(w, r)
	if !ok {
		return
	}
	var req DateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respond(w, 400, ErrorResponse{"Datos de entrada inválidos"})
		return
	}
	result, err := c.service.Preview(r.Context(), id, user, req)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, result)
}

// Create godoc
// @Summary Bloquear por mantenimiento y avisar a los huéspedes
// @Description Suspende las reservas afectadas; el huésped elige reprogramar o solicitar el reembolso total.
// @Tags Mantenimientos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cabinId path int true "ID de cabaña"
// @Param body body CreateRequest true "Rango y motivo del mantenimiento"
// @Success 201 {object} Maintenance
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /maintenance/cabin/{cabinId} [post]
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	id, user, ok := actor(w, r)
	if !ok {
		return
	}
	var req CreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respond(w, 400, ErrorResponse{"Datos de entrada inválidos"})
		return
	}
	result, err := c.service.Create(r.Context(), id, user, req)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 201, result)
}

// Cancel godoc
// @Summary Cancelar un mantenimiento y liberar sus fechas
// @Description Restituye las reservas pendientes cuyas fechas originales siguen disponibles. No revierte las decisiones ya realizadas.
// @Tags Mantenimientos
// @Produce json
// @Security BearerAuth
// @Param cabinId path int true "ID de cabaña"
// @Param id path int true "ID de mantenimiento"
// @Success 200 {object} Maintenance
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /maintenance/cabin/{cabinId}/{id}/cancel [patch]
func (c *Controller) Cancel(w http.ResponseWriter, r *http.Request) {
	cabin, user, ok := actor(w, r)
	if !ok {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond(w, 400, ErrorResponse{"ID de mantenimiento inválido"})
		return
	}
	result, err := c.service.Cancel(r.Context(), cabin, id, user)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, result)
}

func guestActor(w http.ResponseWriter, r *http.Request) (int, uint, bool) {
	user, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond(w, 401, ErrorResponse{"Usuario no autenticado"})
		return 0, 0, false
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond(w, 400, ErrorResponse{"ID de reserva inválido"})
		return 0, 0, false
	}
	return id, user, true
}

// Pending godoc
// @Summary Consultar las opciones de una reserva afectada propia
// @Tags Mantenimientos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de reserva"
// @Success 200 {object} PendingDecision
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /maintenance/reservation/{id} [get]
func (c *Controller) Pending(w http.ResponseWriter, r *http.Request) {
	id, user, ok := guestActor(w, r)
	if !ok {
		return
	}
	result, err := c.service.Pending(r.Context(), id, user)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, result)
}

// Resolve godoc
// @Summary El huésped reprograma en la misma cabaña o pide el reembolso total
// @Description Conserva las noches y el total originales al reprogramar. Solo puede decidir el propietario de la reserva.
// @Tags Mantenimientos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de reserva"
// @Param body body DecisionRequest true "Decisión del huésped"
// @Success 200 {object} DecisionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /maintenance/reservation/{id}/resolve [post]
func (c *Controller) Resolve(w http.ResponseWriter, r *http.Request) {
	id, user, ok := guestActor(w, r)
	if !ok {
		return
	}
	var req DecisionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respond(w, 400, ErrorResponse{"Datos de entrada inválidos"})
		return
	}
	if err := c.service.Resolve(r.Context(), id, user, req); err != nil {
		fail(w, err)
		return
	}
	message := "Fechas actualizadas en la misma cabaña; se conservan el total original y los pagos"
	if req.Action == "reembolsar" {
		message = "Reserva cancelada y reembolso total registrado sin penalizaciones"
	}
	respond(w, 200, DecisionResponse{message})
}
