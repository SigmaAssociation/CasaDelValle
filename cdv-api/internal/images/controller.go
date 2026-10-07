package images

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

func pathIntParam(r *http.Request, name string) (int, bool) {
	value := r.PathValue(name)
	if value == "" {
		value = r.URL.Query().Get(name)
	}

	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

// GetImagesByCabin lista todas las imágenes de una cabaña.
func (c *Controller) GetImagesByCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	cabinID, ok := pathIntParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	result, err := c.service.GetImagesByCabin(r.Context(), cabinID)
	if err != nil {
		if err.Error() == "ID de cabaña inválido" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "Error al obtener las imágenes de la cabaña")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetMainImageByCabin retorna la imagen más reciente de una cabaña.
func (c *Controller) GetMainImageByCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	cabinID, ok := pathIntParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	image, err := c.service.GetMainImageByCabin(r.Context(), cabinID)
	if err != nil {
		switch err.Error() {
		case "ID de cabaña inválido":
			writeError(w, http.StatusBadRequest, err.Error())
		case "La cabaña no tiene imágenes registradas":
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "Error al obtener la imagen principal de la cabaña")
		}
		return
	}

	writeJSON(w, http.StatusOK, image.ToResponse())
}

// GetImageByID retorna una imagen por su identificador.
func (c *Controller) GetImageByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id, ok := pathIntParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "ID de imagen inválido")
		return
	}

	image, err := c.service.GetImageByID(r.Context(), id)
	if err != nil {
		switch err.Error() {
		case "ID de imagen inválido":
			writeError(w, http.StatusBadRequest, err.Error())
		case "Imagen no encontrada":
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "Error al obtener la imagen")
		}
		return
	}

	writeJSON(w, http.StatusOK, image.ToResponse())
}
