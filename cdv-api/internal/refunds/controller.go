package refunds

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"cdv-api/internal/middleware"
	"cdv-api/internal/reservations"
)

type ReservationLookup interface {
	GetReservationByID(ctx context.Context, id int) (*reservations.ReservationDetail, error)
}

type Controller struct {
	service      *Service
	reservations ReservationLookup
}

func NewController(service *Service, reservations ReservationLookup) *Controller {
	return &Controller{
		service:      service,
		reservations: reservations,
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

func (c *Controller) CreateRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id, err := parseReservationID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de reservación inválido")
		return
	}

	reservation, err := c.reservations.GetReservationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, reservations.ErrReservationNotFound) {
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la reservación")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, reservation.HostID) {
		writeError(w, http.StatusForbidden, "No tiene permisos para registrar un reembolso de esta reservación")
		return
	}

	var req CreateRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}

	refund, err := c.service.CreateRefund(r.Context(), reservation, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrReservationNotRefundable):
			writeError(w, http.StatusConflict, "Solo se puede registrar un reembolso para una reservación cancelada")
		case errors.Is(err, ErrNothingRefundable):
			writeError(w, http.StatusConflict, "La reservación no tiene un monto reembolsable disponible")
		case errors.Is(err, ErrRefundExceedsAvailable):
			writeError(w, http.StatusConflict, "El monto del reembolso supera el monto reembolsable disponible")
		case errors.Is(err, ErrRefundInvalidAmount):
			writeError(w, http.StatusBadRequest, "El monto del reembolso debe ser mayor a cero")
		case errors.Is(err, ErrReservationNotFound):
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
		default:
			writeError(w, http.StatusInternalServerError, "Error al registrar el reembolso")
		}
		return
	}

	writeJSON(w, http.StatusCreated, CreateRefundResponse{
		Message: "Reembolso registrado exitosamente",
		Refund:  refund,
	})
}

func (c *Controller) GetRefunds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id, err := parseReservationID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de reservación inválido")
		return
	}

	reservation, err := c.reservations.GetReservationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, reservations.ErrReservationNotFound) {
			writeError(w, http.StatusNotFound, "Reservación no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la reservación")
		return
	}

	canView := middleware.AuthorizeSelfOrAdmin(r, reservation.UserID) ||
		middleware.AuthorizeSelfOrAdmin(r, reservation.HostID)
	if !canView {
		writeError(w, http.StatusForbidden, "No tiene permisos para consultar el historial financiero de esta reservación")
		return
	}

	history, err := c.service.GetHistory(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener el historial financiero")
		return
	}

	writeJSON(w, http.StatusOK, history)
}

func parseReservationID(r *http.Request) (int, error) {
	idParam := r.PathValue("id")
	if idParam == "" {
		idParam = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
