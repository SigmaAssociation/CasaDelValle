package models

type RegisterRequest struct {
	Name     string `json:"nombre"`
	DPI      string `json:"dpi"`
	Email    string `json:"correo"`
	Password string `json:"contrasena"`
	IDRole   int    `json:"id_rol"`
}

type RegisterResponse struct {
	Message string `json:"mensaje"`
	UserID  int    `json:"user_id,omitempty"`
}

type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Adress   string `json:"adress"`
	DPI      string `json:"dpi"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IDRole   uint   `json:"id_role"`
}
