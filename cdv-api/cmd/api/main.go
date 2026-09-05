package main

import (
	"context"
	"log"
	"net/http"

	"cdv-api/internal/database"
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
	mux.HandleFunc(
		"GET /cdv-api/users",
		userController.GetUsers,
	)

	// -------------------------
	// Server
	// -------------------------

	log.Println(
		"Servidor escuchando en http://localhost:8080",
	)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
