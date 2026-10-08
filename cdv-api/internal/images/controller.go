package images

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

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

// CreateImage sube una imagen a una cabaña. Solo el propietario o un admin.
func (c *Controller) CreateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	cabinID, ok := pathIntParam(r, "cabinId")
	if !ok {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	// Obtener hostID de la cabaña
	hostID, err := c.service.repository.GetCabinHostID(r.Context(), cabinID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al verificar la cabaña")
		return
	}

	// Autorización: solo dueño de la cabaña o admin
	if !middleware.AuthorizeSelfOrAdmin(r, uint(hostID)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para agregar imágenes a esta cabaña")
		return
	}

	// Parse multipart (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Error al procesar el formulario multipart")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Campo 'image' requerido en form-data")
		return
	}
	defer file.Close()

	// Obtener userID del token
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	// Crear imagen
	imageID, path, err := c.service.CreateImage(r.Context(), cabinID, int(userID), file, header)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "Cabaña no encontrada" || err.Error() == "La cabaña especificada no existe" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CreateImageResponse{
		Message: "Imagen subida exitosamente",
		ImageID: imageID,
		Path:    path,
	})
}

// GetImagesByCabin lista todas las imágenes de una cabaña.
func (c *Controller) GetImagesByCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	cabinID, ok := pathIntParam(r, "cabinId")
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

	cabinID, ok := pathIntParam(r, "cabinId")
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

// DeleteImage elimina una imagen de una cabaña. Solo el propietario o un admin.
func (c *Controller) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id, ok := pathIntParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "ID de imagen inválido")
		return
	}

	img, err := c.service.GetImageByID(r.Context(), id)
	if err != nil {
		if err.Error() == "Imagen no encontrada" {
			writeError(w, http.StatusNotFound, "Imagen no encontrada")
			return
		}
		if err.Error() == "ID de imagen inválido" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la imagen")
		return
	}

	// Autorización: solo el dueño de la cabaña de la imagen o un admin
	hostID, err := c.service.repository.GetCabinHostID(r.Context(), cabinIDOf(img))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al verificar la cabaña")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(hostID)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para eliminar esta imagen")
		return
	}

	rows, err := c.service.DeleteImage(r.Context(), img)
	if err != nil {
		if err.Error() == "Imagen no encontrada" {
			writeError(w, http.StatusNotFound, "Imagen no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == 0 {
		writeError(w, http.StatusNotFound, "Imagen no encontrada")
		return
	}

	writeJSON(w, http.StatusOK, DeleteImageResponse{
		Message:      "Imagen eliminada exitosamente",
		RowsAffected: rows,
	})
}

func cabinIDOf(img Image) int {
	if img.CabinID == nil {
		return 0
	}
	return *img.CabinID
}
