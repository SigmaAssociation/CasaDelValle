package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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

	// -------------------------
	// Router
	// -------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cdv-api/users", userController.GetUsers)
	mux.HandleFunc("POST /cdv-api/login", userController.LoginUser)
	mux.HandleFunc("POST /cdv-api/users", userController.RegisterUser)
	mux.HandleFunc("GET /cdv-api/users/{id}", userController.GetUserByID)
	mux.HandleFunc("PUT /cdv-api/users/{id}", userController.UpdateUser)

	// -------------------------
	// Server
	// -------------------------

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	handler := middleware.CORS(mux)

	log.Printf("Servidor escuchando en %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
