package reportreservations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cdv-api/internal/middleware"
)

const formatoFecha = "2006-01-02"

var zonaGT = func() *time.Location {
	if loc, err := time.LoadLocation("America/Guatemala"); err == nil {
		return loc
	}
	return time.FixedZone("America/Guatemala", -6*60*60)
}()

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
// @Description Devuelve las cabañas ordenadas por cantidad de reservaciones (descendente), con la información de cada reservación. Solo aparecen cabañas con al menos una reservación que cumpla los filtros. Las fechas se interpretan en hora de Guatemala. Solo administradores. Requiere token.
// @Tags Reportes
// @Produce json
// @Security BearerAuth
// @Param host_id query int false "Filtro por ID de anfitrión"
// @Param cabin_id query int false "Filtro por ID de cabaña"
// @Param guest_id query int false "Filtro por ID de huésped"
// @Param status query string false "Estado de la reservación: pendiente, confirmada, en_curso, finalizada, cancelada"
// @Param created_from query string false "Reservaciones creadas desde esta fecha, inclusive (YYYY-MM-DD)"
// @Param created_to query string false "Reservaciones creadas hasta esta fecha, inclusive (YYYY-MM-DD)"
// @Success 200 {object} ReservationsReport "Reporte de reservaciones"
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

	guestID, err := parseOptionalInt(r, "guest_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdFrom, err := parseOptionalDate(r, "created_from")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdTo, err := parseOptionalDate(r, "created_to")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))

	report, err := c.service.GenerateReport(r.Context(), ReportFilter{
		HostID:      hostID,
		CabinID:     cabinID,
		GuestID:     guestID,
		Status:      status,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidFilter) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al generar el reporte de reservaciones")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

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

func parseOptionalDate(r *http.Request, key string) (*time.Time, error) {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return nil, nil
	}

	t, err := time.ParseInLocation(formatoFecha, v, zonaGT)
	if err != nil {
		return nil, errors.New(key + " debe tener el formato YYYY-MM-DD")
	}

	return &t, nil
}