package seasonprices

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

// DeleteSeasonPrice elimina una tarifa personalizada o una temporada personalizada de una cabaña.
// @Summary Eliminar tarifa o temporada personalizada
// @Description Elimina un registro de precios_temporada de la cabaña indicada. Si era una tarifa sobre una temporada global, la cabaña vuelve al ajuste sugerido; si era una temporada propia, esas fechas vuelven al precio normal. No modifica las temporadas globales. Solo el anfitrión de la cabaña o un admin. Requiere token.
// @Tags Precios de temporada
// @Produce json
// @Security BearerAuth
// @Param cabinId path int true "ID de cabaña"
// @Param id path int true "ID de precio de temporada"
// @Success 200 {object} DeleteSeasonPriceResponse "Tarifa eliminada exitosamente"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Cabaña o tarifa no encontrada"
// @Router /cabins/{cabinId}/season-prices/{id} [delete]
func (c *Controller) DeleteSeasonPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	cabinID, err := strconv.ParseUint(r.PathValue("cabinId"), 10, 64)
	if err != nil || cabinID == 0 {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de tarifa inválido")
		return
	}

	hostID, err := c.service.GetCabinHostID(r.Context(), uint(cabinID))
	if err != nil {
		if errors.Is(err, ErrCabinNotFound) {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al eliminar la tarifa")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, hostID) {
		writeError(w, http.StatusForbidden, "No tiene permisos para eliminar esta tarifa")
		return
	}

	if err := c.service.Delete(r.Context(), uint(cabinID), id); err != nil {
		switch {
		case errors.Is(err, ErrSeasonPriceNotFound):
			writeError(w, http.StatusNotFound, "Tarifa no encontrada")
		default:
			writeError(w, http.StatusInternalServerError, "Error al eliminar la tarifa")
		}
		return
	}

	writeJSON(w, http.StatusOK, DeleteSeasonPriceResponse{
		Message: "Tarifa eliminada exitosamente",
	})
}