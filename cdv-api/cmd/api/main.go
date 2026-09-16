package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cdv-api/internal/cabins"
	"cdv-api/internal/database"
	"cdv-api/internal/middleware"
	"cdv-api/internal/users"
)

func main() {

	// -------------------------
	// Database
	// -------------------------
	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx)

	if err != nil {
		log.Fatalf(
			"Error conectando a PostgreSQL: %v",
			err,
		)
	}

	defer pool.Close()

	log.Println("PostgreSQL conectado")

	// -------------------------
	// Dependency Injection
	// -------------------------
	userRepository := users.NewRepository(pool)
	userService := users.NewService(userRepository)
	userController := users.NewController(userService)

	cabinRepository := cabins.NewRepository(pool)
	cabinService := cabins.NewService(cabinRepository)
	cabinController := cabins.NewController(cabinService)

	// -------------------------
	// Router
	// -------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cdv-api/users", userController.GetUsers)
	mux.HandleFunc("POST /cdv-api/login", userController.LoginUser)
	mux.HandleFunc("POST /cdv-api/users", userController.RegisterUser)
	mux.HandleFunc("GET /cdv-api/users/{id}", userController.GetUserByID)
	mux.HandleFunc("PUT /cdv-api/users/{id}", userController.UpdateUser)

	mux.HandleFunc("GET /cdv-api/cabins", cabinController.GetCabins)
	mux.HandleFunc("POST /cdv-api/cabins", cabinController.CreateCabin)
	mux.HandleFunc("GET /cdv-api/cabins/user/{userId}", cabinController.GetCabinsByUser)
	mux.HandleFunc("GET /cdv-api/cabins/{id}", cabinController.GetCabinByID)
	mux.HandleFunc("DELETE /cdv-api/cabins/{id}", cabinController.DeleteCabin)

	// -------------------------
	// Server
	// -------------------------
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	handler := middleware.CORS(middleware.Auth(mux))

	log.Printf("Servidor escuchando en %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
