package users

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
