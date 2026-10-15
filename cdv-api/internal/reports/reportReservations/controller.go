package reportreservations

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

// GetReservationsReport obtiene el reporte de reservaciones por cabaña.
// @Summary Reporte de reservaciones por cabaña
// @Description Devuelve las cabañas ordenadas por cantidad de reservaciones (descendente), con la información de cada reservación. Filtros opcionales por anfitrión y cabaña. Solo administradores. Requiere token.
// @Tags Reportes
// @Produce json
// @Security BearerAuth
// @Param host_id query int false "Filtro por ID de anfitrión"
// @Param cabin_id query int false "Filtro por ID de cabaña"
// @Success 200 {object} ReporteReservaciones "Reporte de reservaciones"
// @Failure 400 {object} ErrorResponse "Parámetros inválidos"
// @Failure 403 {object} ErrorResponse "Sin permisos"
// @Failure 500 {object} ErrorResponse "Error al generar el reporte"
// @Router /reports/reservations [get]
func (c *Controller) GetReservationsReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	if !middleware.IsAdmin(r) {
		writeError(w, http.StatusForbidden, "Solo los administradores pueden ver este reporte")
		return
	}

	hostID, err := parseOptionalInt(r, "host_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cabinID, err := parseOptionalInt(r, "cabin_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	report, err := c.service.GenerarReporte(r.Context(), FiltroReporte{
		IDAnfitrion: hostID,
		IDCabana:    cabinID,
	})
	if err != nil {
		if errors.Is(err, ErrFiltroInvalido) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al generar el reporte de reservaciones")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// parseOptionalInt convierte un query param opcional a *int.
// Solo verifica el formato; las reglas de negocio las valida el service.
func parseOptionalInt(r *http.Request, key string) (*int, error) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil, nil
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, errors.New(key + " debe ser un número entero válido")
	}

	return &n, nil
}