package seasons

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
	return &Controller{service: service}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}

// ListSeasons lista las temporadas globales
// @Summary Listar temporadas
// @Tags Temporadas
// @Accept json
// @Produce json
// @Param all query boolean false "Incluir inactivas (solo admin)"
// @Success 200 {array} Season
// @Router /seasons [get]
func (c *Controller) ListSeasons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	includeInactive := r.URL.Query().Get("all") == "true" && middleware.IsAdmin(r)
	seasons, err := c.service.ListSeasons(r.Context(), includeInactive)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener temporadas")
		return
	}
	writeJSON(w, http.StatusOK, seasons)
}

// CreateSeason crea una temporada global
// @Summary Crear temporada
// @Tags Temporadas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateSeasonRequest true "Datos de la temporada"
// @Success 201 {object} Season
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /seasons [post]
func (c *Controller) CreateSeason(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	if !middleware.IsAdmin(r) {
		writeError(w, http.StatusForbidden, "Solo administradores pueden crear temporadas")
		return
	}
	var req CreateSeasonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}
	season, err := c.service.CreateSeason(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrSeasonOverlap) {
			writeError(w, http.StatusConflict, "Ya existe una temporada con fechas cruzadas")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, season)
}

// UpdateSeason actualiza una temporada global
// @Summary Actualizar temporada
// @Tags Temporadas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de la temporada"
// @Param body body UpdateSeasonRequest true "Datos de la temporada"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /seasons/{id} [put]
func (c *Controller) UpdateSeason(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	if !middleware.IsAdmin(r) {
		writeError(w, http.StatusForbidden, "Solo administradores pueden actualizar temporadas")
		return
	}
	idParam := r.PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}
	var req UpdateSeasonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}
	if err := c.service.UpdateSeason(r.Context(), id, req); err != nil {
		if errors.Is(err, ErrSeasonNotFound) {
			writeError(w, http.StatusNotFound, "Temporada no encontrada")
			return
		}
		if errors.Is(err, ErrSeasonOverlap) {
			writeError(w, http.StatusConflict, "Ya existe una temporada con fechas cruzadas")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Temporada actualizada"})
}

// DeleteSeason elimina una temporada global
// @Summary Eliminar temporada
// @Tags Temporadas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de la temporada"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{id} [delete]
func (c *Controller) DeleteSeason(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	if !middleware.IsAdmin(r) {
		writeError(w, http.StatusForbidden, "Solo administradores pueden eliminar temporadas")
		return
	}
	idParam := r.PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}
	if err := c.service.DeleteSeason(r.Context(), id); err != nil {
		if errors.Is(err, ErrSeasonNotFound) {
			writeError(w, http.StatusNotFound, "Temporada no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al eliminar temporada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Temporada eliminada"})
}

// ListCabinSeasonPrices lista precios por temporada de una cabaña
// @Summary Listar precios por temporada de una cabaña
// @Tags Precios por temporada
// @Accept json
// @Produce json
// @Param cabinId path int true "ID de la cabaña"
// @Success 200 {array} CabinSeasonPrice
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cabin-season-prices/cabin/{cabinId} [get]
func (c *Controller) ListCabinSeasonPrices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	cabinID, err := strconv.ParseUint(r.PathValue("cabinId"), 10, 64)
	if err != nil || cabinID == 0 {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}
	items, err := c.service.ListCabinSeasonPrices(r.Context(), uint(cabinID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener precios por temporada")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// SetCabinSeasonPrice configura precio por temporada para una cabaña
// @Summary Configurar precio por temporada
// @Tags Precios por temporada
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cabinId path int true "ID de la cabaña"
// @Param body body SetCabinSeasonPriceRequest true "Configuración de precio"
// @Success 201 {object} CabinSeasonPrice
// @Failure 400 {object} ErrorResponse
// @Router /cabin-season-prices/cabin/{cabinId} [post]
func (c *Controller) SetCabinSeasonPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	cabinID, err := strconv.ParseUint(r.PathValue("cabinId"), 10, 64)
	if err != nil || cabinID == 0 {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}
	var req SetCabinSeasonPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}
	item, err := c.service.SetCabinSeasonPrice(r.Context(), uint(cabinID), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

// DeleteCabinSeasonPrice elimina una configuración de precio por temporada
// @Summary Eliminar configuración de precio por temporada
// @Tags Precios por temporada
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de la configuración"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cabin-season-prices/{id} [delete]
func (c *Controller) DeleteCabinSeasonPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}
	if err := c.service.DeleteCabinSeasonPrice(r.Context(), id); err != nil {
		if errors.Is(err, ErrCabinSeasonNotFound) {
			writeError(w, http.StatusNotFound, "Configuración no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al eliminar configuración")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Configuración eliminada"})
}

// GetPriceQuote cotiza el precio por fecha con ajuste por temporada
// @Summary Cotizar precio por temporada
// @Tags Precios por temporada
// @Accept json
// @Produce json
// @Param body body PriceQuoteRequest true "Datos para cotización"
// @Success 200 {object} PriceQuoteResponse
// @Failure 400 {object} ErrorResponse
// @Router /seasons/quote [post]
func (c *Controller) GetPriceQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	var req PriceQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}
	quote, err := c.service.GetPriceQuote(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidDates) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}
