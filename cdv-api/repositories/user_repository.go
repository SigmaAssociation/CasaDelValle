package repositories

import "cdv-api/models"

type UserRepository struct {}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) GetUsers() ([]models.User, error) {
	return []models.User{
		{ID: 1, Name: "Mynor God", Phone: "1234567890", Adress: "123 Main St", DPI: "1234567890123", Email: "m1@example.com", Password: "password", IDRole: 1},	
		{ID: 2, Name: "MauLan Ixcoy", Phone: "0987654321", Adress: "456 Elm St", DPI: "9876543210987", Email: "m2@example.com", Password: "password", IDRole: 2},
		{ID: 3, Name: "Mane Grubi", Phone: "5555555555", Adress: "789 Oak St", DPI: "5555555555555", Email: "m3@example.com", Password: "password", IDRole: 3},
	}, nil
}