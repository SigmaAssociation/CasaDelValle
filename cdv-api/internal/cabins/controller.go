package cabins

import (
	"encoding/json"
	"net/http"
	"strconv"
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
