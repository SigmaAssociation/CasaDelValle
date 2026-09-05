package users

type RegisterRequest struct {
	Name     string `json:"name"`
	DPI      string `json:"dpi"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Message string `json:"mensaje"`
	UserID  int    `json:"user_id,omitempty"`
}

type User struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	DPI     string `json:"dpi"`
	Email   string `json:"email"`
	IDRole  uint   `json:"id_role"`
}
