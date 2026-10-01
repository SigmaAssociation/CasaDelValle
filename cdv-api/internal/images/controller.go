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
	return &Controller{service: service}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Message: message})
}

func (c *Controller) CreateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	// cabinId de la URL
	idParam := r.PathValue("cabinId")
	if idParam == "" {
		idParam = r.URL.Query().Get("cabinId")
	}
	cabinID, err := strconv.Atoi(idParam)
	if err != nil || cabinID <= 0 {
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
	imageID, ruta, err := c.service.CreateImage(r.Context(), cabinID, int(userID), file, header)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "Cabaña no encontrada" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CreateImageResponse{
		Message: "Imagen subida exitosamente",
		ImageID: imageID,
		Ruta:    ruta,
	})
}

func (c *Controller) GetImagesByCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("cabinId")
	if idParam == "" {
		idParam = r.URL.Query().Get("cabinId")
	}
	cabinID, err := strconv.Atoi(idParam)
	if err != nil || cabinID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	images, err := c.service.GetImagesByCabin(r.Context(), cabinID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener imágenes")
		return
	}

	writeJSON(w, http.StatusOK, ImagesListResponse{
		Data:  images,
		Total: len(images),
	})
}

func (c *Controller) DeleteImage(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusBadRequest, "ID de imagen inválido")
		return
	}

	img, err := c.service.GetImageByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Imagen no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener la imagen")
		return
	}

	// Autorización: solo el dueño de la cabaña de la imagen o un admin
	hostID, err := c.service.repository.GetCabinHostID(r.Context(), img.IDCabana)
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == 0 {
		writeError(w, http.StatusNotFound, "Imagen no encontrada")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"mensaje":       "Imagen eliminada exitosamente",
		"rows_affected": rows,
	})
}