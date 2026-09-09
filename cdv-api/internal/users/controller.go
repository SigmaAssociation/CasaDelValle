package users

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

func (c *Controller) GetUsers(w http.ResponseWriter, r *http.Request) {
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

/*
Intenta autenticar a un usuario con el correo y la contraseña proporcionados en la solicitud.
Si la autenticación es exitosa, devuelve un mensaje y el token JWT. Si falla, devuelve un mensaje de error.
*/
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

func (c *Controller) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Error(w, "ID de usuario no proporcionado", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
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
