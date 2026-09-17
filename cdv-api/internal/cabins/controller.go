package cabins

import (
	"encoding/json"
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

func (c *Controller) GetCabins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	cabins, err := c.service.GetCabins(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener las cabañas")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(cabins))
}

func (c *Controller) CreateCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var req CreateCabinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(req.IDAnfitrion)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para registrar esta cabaña")
		return
	}

	cabinID, err := c.service.CreateCabin(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CreateCabinResponse{
		Message: "Cabaña registrada exitosamente",
		CabinID: cabinID,
	})
}

func (c *Controller) DeleteCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		idParam = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	cabin, err := c.service.GetCabinByID(r.Context(), GetCabinByIDRequest{ID: id})
	if err != nil {
		if err.Error() == "Cabaña no encontrada" {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la cabaña")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(cabin.HostID)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para eliminar esta cabaña")
		return
	}

	rowsAffected, err := c.service.DeleteCabin(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, DeleteCabinResponse{
		Message:      "Cabaña eliminada exitosamente",
		RowsAffected: rowsAffected,
	})
}

func (c *Controller) GetCabinByID(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	cabin, err := c.service.GetCabinByID(r.Context(), GetCabinByIDRequest{ID: id})
	if err != nil {
		if err.Error() == "Cabaña no encontrada" {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}

		if err.Error() == "ID de cabaña inválido" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "Error al obtener la cabaña")
		return
	}

	writeJSON(w, http.StatusOK, cabin.ToResponse())
}

func (c *Controller) GetCabinsByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("userId")
	if idParam == "" {
		idParam = r.URL.Query().Get("userId")
	}

	userID, err := strconv.Atoi(idParam)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de usuario inválido")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(userID)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para ver las cabañas de este usuario")
		return
	}

	cabins, err := c.service.GetCabinsByUser(
		r.Context(),
		GetCabinsByUserRequest{UserID: userID},
	)
	if err != nil {
		if err.Error() == "ID de usuario inválido" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "Error al obtener las cabañas del usuario")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(cabins))
}

/*BUSQUEDA POR FILTROS COMBINADOS*/
func (c *Controller) SearchCabins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	query := r.URL.Query()
	params := CabinSearchParams{}

	if v := query.Get("host_id"); v != "" {
		hostID, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "host_id debe ser un número entero válido")
			return
		}
		params.HostID = &hostID
	}

	if v := query.Get("min_capacity"); v != "" {
		minCap, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "min_capacity debe ser un número entero válido")
			return
		}
		params.MinCapacity = &minCap
	}

	if v := query.Get("max_capacity"); v != "" {
		maxCap, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "max_capacity debe ser un número entero válido")
			return
		}
		params.MaxCapacity = &maxCap
	}

	if v := query.Get("min_price"); v != "" {
		minPrice, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "min_price debe ser un número válido")
			return
		}
		params.MinPrice = &minPrice
	}

	if v := query.Get("max_price"); v != "" {
		maxPrice, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "max_price debe ser un número válido")
			return
		}
		params.MaxPrice = &maxPrice
	}

	result, err := c.service.SearchCabins(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (c *Controller) UpdateCabin(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	var req UpdateCabinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}

	existing, err := c.service.GetCabinByID(r.Context(), GetCabinByIDRequest{ID: id})
	if err != nil {
		if err.Error() == "Cabaña no encontrada" {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la cabaña")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(existing.HostID)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para editar esta cabaña")
		return
	}

	if err := c.service.UpdateCabin(r.Context(), id, req); err != nil {
		if err.Error() == "Cabaña no encontrada" {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"mensaje": "Cabaña actualizada exitosamente"})
}
