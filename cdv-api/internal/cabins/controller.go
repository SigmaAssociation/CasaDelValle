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

// GetCabins obtiene todas las cabañas registradas.
// @Summary Listar cabañas
// @Description Devuelve el listado completo de cabañas.
// @Tags Cabañas
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CabinsListResponse "Lista de cabañas"
// @Failure 500 {object} map[string]string "Error al obtener las cabañas"
// @Router /cabins [get]
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

// CreateCabin registra una nueva cabaña.
// @Summary Registrar cabaña
// @Description Crea una cabaña. Para anfitriones el id_anfitrion se toma del token; para administradores puede indicarse. Requiere token.
// @Tags Cabañas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cabin body CreateCabinRequest true "Datos de la cabaña"
// @Success 201 {object} CreateCabinResponse "Cabaña registrada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Router /cabins [post]
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

	userIDUint, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	if !middleware.IsAdmin(r) {
		req.HostID = int(userIDUint)
	} else if req.HostID == 0 {
		req.HostID = int(userIDUint)
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(req.HostID)) {
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

// DeleteCabin elimina una cabaña por su ID.
// @Summary Eliminar cabaña
// @Description Elimina una cabaña. Solo el dueño de la cabaña o un administrador. Requiere token.
// @Tags Cabañas
// @Security BearerAuth
// @Param id path int true "ID de cabaña"
// @Success 200 {object} DeleteCabinResponse "Cabaña eliminada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Cabaña no encontrada"
// @Router /cabins/{id} [delete]
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

// GetCabinByID obtiene los detalles de una cabaña.
// @Summary Obtener cabaña por ID
// @Description Devuelve los detalles de la cabaña indicada. Requiere token.
// @Tags Cabañas
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de cabaña"
// @Success 200 {object} CabinResponse "Detalle de la cabaña"
// @Failure 400 {object} map[string]string "ID de cabaña inválido"
// @Failure 404 {object} map[string]string "Cabaña no encontrada"
// @Router /cabins/{id} [get]
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

// GetCabinsByUser obtiene las cabañas de un anfitrión.
// @Summary Cabañas de un usuario
// @Description Devuelve las cabañas cuyo anfitrión es el usuario indicado. Solo el propio usuario o un administrador. Requiere token.
// @Tags Cabañas
// @Produce json
// @Security BearerAuth
// @Param userId path int true "ID del anfitrión"
// @Success 200 {object} CabinsListResponse "Cabañas del usuario"
// @Failure 400 {object} map[string]string "ID de usuario inválido"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Router /cabins/user/{userId} [get]
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

// SearchCabins busca cabañas con filtros combinados.
// @Summary Buscar cabañas
// @Description Busca cabañas combinando filtros opcionales por nombre, anfitrión, capacidad y precio. Requiere token.
// @Tags Cabañas
// @Produce json
// @Security BearerAuth
// @Param name query string false "Filtro por nombre"
// @Param host_id query int false "Filtro por ID de anfitrión"
// @Param host_name query string false "Filtro por nombre de anfitrión"
// @Param min_capacity query int false "Capacidad mínima"
// @Param max_capacity query int false "Capacidad máxima"
// @Param min_price query number false "Precio mínimo"
// @Param max_price query number false "Precio máximo"
// @Success 200 {object} CabinCardsListResponse "Resultados de la búsqueda"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Router /cabins/search [get]
func (c *Controller) SearchCabins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	query := r.URL.Query()
	params := CabinSearchParams{}

	if v := query.Get("name"); v != "" {
		params.Name = &v
	}

	if v := query.Get("host_id"); v != "" {
		hostID, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "host_id debe ser un número entero válido")
			return
		}
		params.HostID = &hostID
	}

	if v := query.Get("name"); v != "" {
		params.Name = &v
	}

	if v := query.Get("host_name"); v != "" {
		params.HostName = &v
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

// UpdateCabin actualiza la información de una cabaña.
// @Summary Actualizar cabaña
// @Description Actualiza los datos de una cabaña. Solo el dueño de la cabaña o un administrador. Requiere token.
// @Tags Cabañas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de cabaña"
// @Param cabin body UpdateCabinRequest true "Datos a actualizar"
// @Success 200 {object} map[string]string "Cabaña actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Cabaña no encontrada"
// @Router /cabins/{id} [put]
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

	writeJSON(w, http.StatusOK, map[string]string{"message": "Cabaña actualizada exitosamente"})
}
