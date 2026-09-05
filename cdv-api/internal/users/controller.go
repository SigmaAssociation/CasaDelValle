package users

import (
	"encoding/json"
	"net/http"
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
