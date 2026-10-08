package users

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

// GetUsersList obtiene el listado de usuarios del sistema.
// @Summary Listar usuarios
// @Description Devuelve todos los usuarios registrados. Requiere rol de administrador.
// @Tags Usuarios
// @Produce json
// @Security BearerAuth
// @Success 200 {array} User "Lista de usuarios"
// @Failure 403 {object} map[string]string "Sin permisos de administrador"
// @Failure 500 {object} map[string]string "Error al obtener usuarios"
// @Router /users [get]
func (c *Controller) GetUsers(w http.ResponseWriter, r *http.Request) {
	if !middleware.IsAdmin(r) {
		w.Header().Set(
			"Content-Type",
			"application/json",
		)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "No tiene permisos para ver la lista de usuarios"})
		return
	}

	usuarios, err := c.service.GetUsers(
		r.Context(),
	)

	if err != nil {
		print("Error al obtener usuarios", err)
		http.Error(
			w,
			"Error al obtener usuarios",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(usuarios)
}

// RegisterUser registra un nuevo usuario en el sistema.
// @Summary Registrar usuario
// @Description Crea una cuenta de usuario. Endpoint público, no requiere token.
// @Tags Usuarios
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Datos del usuario"
// @Success 201 {object} RegisterResponse "Usuario registrado exitosamente"
// @Failure 400 {object} map[string]string "Datos de entrada inválidos"
// @Failure 409 {object} map[string]string "El correo o DPI ya está registrado"
// @Router /users [post]
func (c *Controller) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Datos de entrada inválidos", http.StatusBadRequest)
		return
	}
	userID, err := c.service.RegisterUser(r.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "El correo o dpi ya está registrado" {
			status = http.StatusConflict
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(RegisterResponse{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		Message: "Usuario registrado exitosamente",
		UserID:  userID,
	})
}

// UpdateUser actualiza el perfil de un usuario.
// @Summary Actualizar usuario
// @Description Actualiza los datos de perfil. Solo el propio usuario o un administrador. Requiere token.
// @Tags Usuarios
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de usuario"
// @Param user body UserUpdate true "Datos a actualizar"
// @Success 200 {object} map[string]interface{} "Usuario actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos de entrada inválidos"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Usuario no encontrado"
// @Failure 409 {object} map[string]string "El correo ya está registrado por otro usuario"
// @Router /users/{id} [put]
func (c *Controller) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID de usuario inválido en la URL"})
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(userID)) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "No tiene permisos para editar este perfil"})
		return
	}

	var req UserUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Datos de entrada inválidos"})
		return
	}

	req.ID = uint(userID)

	rowsAffected, err := c.service.UpdateUser(r.Context(), req)
	if err != nil {
		status := http.StatusBadRequest

		if err.Error() == "el correo ya está registrado por otro usuario" {
			status = http.StatusConflict
		} else if err.Error() == "Usuario no encontrado" {
			status = http.StatusNotFound
		}

		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Usuario actualizado exitosamente",
		"rows_affected": rowsAffected,
	})
}

// LoginUser autentica a un usuario y devuelve un token JWT.
// @Summary Iniciar sesión
// @Description Autentica con correo y contraseña y devuelve un token Bearer. Endpoint público.
// @Tags Usuarios
// @Accept json
// @Produce json
// @Param credentials body UserLoginRequest true "Credenciales de acceso"
// @Success 200 {object} UserLoginResponse "Usuario autenticado exitosamente"
// @Failure 400 {object} map[string]string "Datos de entrada inválidos"
// @Failure 401 {object} map[string]string "Correo o contraseña incorrectos"
// @Router /login [post]
func (c *Controller) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Datos de entrada inválidos", http.StatusBadRequest)
		return
	}
	token, err := c.service.AuthenticateUser(r.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		message := err.Error()
		if err.Error() == "Error al generar el token" {
			status = http.StatusInternalServerError
			message = "Error interno del servidor"
		}
		if err.Error() == "Contraseña incorrecta" || err.Error() == "Usuario no encontrado" {
			status = http.StatusUnauthorized
			message = "Correo o contraseña incorrectos"
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(UserLoginResponse{Message: message, Token: ""})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserLoginResponse{
		Message: "Usuario autenticado exitosamente",
		Token:   token,
	})
}

// GetUserByID obtiene el perfil de un usuario.
// @Summary Obtener usuario por ID
// @Description Devuelve el perfil de un usuario. Solo el propio usuario o un administrador. Requiere token.
// @Tags Usuarios
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de usuario"
// @Success 200 {object} User "Perfil del usuario"
// @Failure 400 {object} map[string]string "ID de usuario inválido"
// @Failure 403 {object} map[string]string "Sin permisos"
// @Failure 404 {object} map[string]string "Usuario no encontrado"
// @Router /users/{id} [get]
func (c *Controller) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		idParam = r.URL.Query().Get("id")
	}
	if idParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID de usuario no proporcionado"})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID de usuario inválido"})
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(id)) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "No tiene permisos para ver este perfil"})
		return
	}

	user, err := c.service.GetUserByID(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Error al obtener el usuario"
		if err.Error() == "Usuario no encontrado" {
			status = http.StatusNotFound
			message = "Usuario no encontrado"
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"message": message})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
