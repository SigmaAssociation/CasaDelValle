package reportincome

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cdv-api/internal/middleware"
)

var guatemalaZone = func() *time.Location {
	if loc, err := time.LoadLocation("America/Guatemala"); err == nil {
		return loc
	}
	return time.FixedZone("America/Guatemala", -6*60*60)
}()

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Message: message})
}

// GetIncomeReport obtiene el reporte de ingresos por rango de fechas y cabaña.
// @Summary Reporte de ingresos por rango de fechas
// @Description Devuelve los pagos, anticipos y reembolsos registrados en el rango de fechas indicado, junto con el total recaudado neto. Los reembolsos se devuelven con monto negativo. Las fechas se interpretan en hora de Guatemala. Los administradores ven todas las cabañas; los anfitriones solo las propias. Requiere token.
// @Tags Reportes
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Fecha de inicio del rango, inclusive (YYYY-MM-DD)"
// @Param end_date query string true "Fecha de fin del rango, inclusive (YYYY-MM-DD)"
// @Param cabin_id query int false "Filtro por ID de cabaña"
// @Success 200 {object} IncomeReport "Reporte de ingresos"
// @Failure 400 {object} ErrorResponse "Parámetros inválidos"
// @Failure 403 {object} ErrorResponse "Sin permisos"
// @Failure 500 {object} ErrorResponse "Error al generar el reporte"
// @Router /reports/income [get]
func (c *Controller) GetIncomeReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	if !middleware.IsHostOrAdmin(r) {
		writeError(w, http.StatusForbidden, "Solo los administradores y anfitriones pueden ver este reporte")
		return
	}

	startDate, err := parseRequiredDate(r, "start_date")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	endDate, err := parseRequiredDate(r, "end_date")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cabinID, err := parseOptionalInt(r, "cabin_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := ReportFilter{
		StartDate: *startDate,
		EndDate:   *endDate,
		CabinID:   cabinID,
	}

	if role, ok := middleware.RoleFromContext(r.Context()); ok && role == middleware.RoleHost {
		if userID, ok := middleware.UserIDFromContext(r.Context()); ok {
			hostID := userID
			filter.HostID = &hostID
		}
	}

	report, err := c.service.GenerateReport(r.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrInvalidDateRange) || errors.Is(err, ErrInvalidFilter) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al generar el reporte de ingresos")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func parseRequiredDate(r *http.Request, key string) (*time.Time, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return nil, errors.New(key + " es obligatorio y debe tener el formato YYYY-MM-DD")
	}

	parsed, err := time.ParseInLocation(dateLayout, value, guatemalaZone)
	if err != nil {
		return nil, errors.New(key + " debe tener el formato YYYY-MM-DD")
	}

	return &parsed, nil
}

func parseOptionalInt(r *http.Request, key string) (*int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, errors.New(key + " debe ser un número entero válido")
	}

	return &parsed, nil
}
