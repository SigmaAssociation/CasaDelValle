package main

import (
	"cdv-api/controllers"
	"cdv-api/repositories"
	"cdv-api/services"
	"log"
	"net/http"
)

func main() {
	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	http.HandleFunc("/cdv-api/users", userController.GetUsers)

	log.Println("Servidor iniciado en http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}