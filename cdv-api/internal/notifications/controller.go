package notifications

import (
	"encoding/json"
	"errors"
	"log"
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

// GetNotifications devuelve el buzón personal del usuario autenticado, en
// orden cronológico descendente. Admite paginación con limit y offset.
// @Summary Obtener buzón de notificaciones
// @Description Devuelve las notificaciones personales del usuario autenticado en orden cronológico descendente (más reciente primero). El buzón es personal: siempre corresponde al usuario del token. Requiere token.
// @Tags Notificaciones
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Cantidad máxima por página (por defecto 20, máximo 100)"
// @Param offset query int false "Desplazamiento para paginar (por defecto 0)"
// @Success 200 {object} NotificationListResponse "Lista de notificaciones del usuario"
// @Failure 401 {object} map[string]string "Usuario no autenticado"
// @Failure 500 {object} map[string]string "Error interno al obtener las notificaciones"
// @Router /notifications [get]
func (c *Controller) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	limit, offset := parsePagination(r)

	items, total, err := c.service.GetNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		if errors.Is(err, ErrUserRequired) {
			writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
			return
		}
		log.Printf("notifications: error al listar el buzón del usuario %d: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "Error al obtener las notificaciones")
		return
	}

	if items == nil {
		items = []Notification{}
	}

	writeJSON(w, http.StatusOK, NotificationListResponse{
		Data:    items,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: offset+len(items) < total,
	})
}

// parsePagination lee limit y offset desde la query string. Los valores
// inválidos se normalizan a los predeterminados.
func parsePagination(r *http.Request) (int, int) {
	limit, offset := 0, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			offset = parsed
		}
	}
	return NormalizePagination(limit, offset)
}

// GetUnreadCount devuelve cuántas notificaciones sin leer tiene el usuario.
// @Summary Contar notificaciones no leídas
// @Description Devuelve cuántas notificaciones sin leer tiene el usuario autenticado. Requiere token.
// @Tags Notificaciones
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UnreadCountResponse "Cantidad de notificaciones no leídas"
// @Failure 401 {object} map[string]string "Usuario no autenticado"
// @Failure 500 {object} map[string]string "Error interno al contar las notificaciones"
// @Router /notifications/unread-count [get]
func (c *Controller) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	count, err := c.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserRequired) {
			writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
			return
		}
		log.Printf("notifications: error al contar no leídas del usuario %d: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "Error al obtener el contador de notificaciones")
		return
	}

	writeJSON(w, http.StatusOK, UnreadCountResponse{UnreadCount: count})
}

// MarkAllRead marca todas las notificaciones del usuario como leídas.
// @Summary Marcar todas las notificaciones como leídas
// @Description Marca como leídas todas las notificaciones del usuario autenticado. Requiere token.
// @Tags Notificaciones
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MarkAllReadResponse "Resultado de la actualización"
// @Failure 401 {object} map[string]string "Usuario no autenticado"
// @Failure 500 {object} map[string]string "Error interno al actualizar las notificaciones"
// @Router /notifications/read-all [post]
func (c *Controller) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	updated, err := c.service.MarkAllRead(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserRequired) {
			writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
			return
		}
		log.Printf("notifications: error al marcar todo como leído del usuario %d: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "Error al actualizar las notificaciones")
		return
	}

	writeJSON(w, http.StatusOK, MarkAllReadResponse{
		Message: "Notificaciones marcadas como leídas",
		Updated: updated,
	})
}

// MarkRead marca una notificación del usuario autenticado como leída. Por
// privacidad, solo su dueño puede marcarla: si no existe o pertenece a otro
// usuario se devuelve 404.
// @Summary Marcar una notificación como leída
// @Description Marca como leída la notificación indicada. Solo su dueño puede marcarla; si no existe o pertenece a otro usuario se devuelve 404. Requiere token.
// @Tags Notificaciones
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de la notificación"
// @Success 200 {object} MarkReadResponse "Notificación marcada como leída"
// @Failure 400 {object} map[string]string "ID de notificación inválido"
// @Failure 401 {object} map[string]string "Usuario no autenticado"
// @Failure 404 {object} map[string]string "Notificación no encontrada"
// @Router /notifications/{id}/read [patch]
func (c *Controller) MarkRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		idParam = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de notificación inválido")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	if err := c.service.MarkRead(r.Context(), id, userID); err != nil {
		switch {
		case errors.Is(err, ErrNotificationNotFound):
			writeError(w, http.StatusNotFound, "Notificación no encontrada")
		case errors.Is(err, ErrUserRequired):
			writeError(w, http.StatusUnauthorized, "Usuario no autenticado")
		case err.Error() == "ID de notificación inválido":
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("notifications: error al marcar la notificación %d del usuario %d: %v", id, userID, err)
			writeError(w, http.StatusInternalServerError, "Error al actualizar la notificación")
		}
		return
	}

	writeJSON(w, http.StatusOK, MarkReadResponse{
		Message: "Notificación marcada como leída",
	})
}
