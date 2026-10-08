// Casa Del Valle API
// @title Casa Del Valle API
// @version 1.0
// @description Documentación de los endpoints de la API de Casa Del Valle.
// @host localhost:8080
// @BasePath /cdv-api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Token JWT obtenido en el endpoint login. Enviar como: Bearer <token>
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"cdv-api/internal/cabins"
	"cdv-api/internal/database"
	"cdv-api/internal/images"
	"cdv-api/internal/middleware"
	"cdv-api/internal/reservations"
	"cdv-api/internal/users"

	_ "cdv-api/docs"
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

	reservationRepository := reservations.NewRepository(pool)
	reservationService := reservations.NewService(reservationRepository)
	reservationController := reservations.NewController(reservationService)

	imageRepository := images.NewRepository(pool)
	imageService := images.NewService(imageRepository)
	imageController := images.NewController(imageService)

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

	mux.HandleFunc("GET /cdv-api/cabins/search", cabinController.SearchCabins)
	mux.HandleFunc("GET /cdv-api/cabins/user/{userId}", cabinController.GetCabinsByUser)

	mux.HandleFunc("GET /cdv-api/cabins/{id}", cabinController.GetCabinByID)
	mux.HandleFunc("PUT /cdv-api/cabins/{id}", cabinController.UpdateCabin)
	mux.HandleFunc("DELETE /cdv-api/cabins/{id}", cabinController.DeleteCabin)

	mux.HandleFunc("POST /cdv-api/cabins/images/{cabinId}", imageController.CreateImage)
	mux.HandleFunc("GET /cdv-api/cabins/images/{cabinId}", imageController.GetImagesByCabin)
	mux.HandleFunc("GET /cdv-api/cabins/images/{cabinId}/principal", imageController.GetMainImageByCabin)
	mux.HandleFunc("GET /cdv-api/images/{id}", imageController.GetImageByID)
	mux.HandleFunc("DELETE /cdv-api/images/{id}", imageController.DeleteImage)

	mux.HandleFunc("POST /cdv-api/reservations", reservationController.CreateReservation)
	mux.HandleFunc("PUT /cdv-api/reservations/{id}", reservationController.UpdateReservation)
	mux.HandleFunc("PATCH /cdv-api/reservations/{id}/cancel", reservationController.CancelReservation)
	// -------------------------
	// Server
	// -------------------------
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)

	// Servir archivos subidos (imágenes de cabañas) de forma pública, para que
	// cualquier usuario pueda verlas como referencia visual del catálogo.
	mux.Handle(
		"/cdv-api/uploads/",
		http.StripPrefix("/cdv-api/uploads/", http.FileServer(http.Dir("./uploads"))),
	)

	// Swagger UI: documentación interactiva de la API (acceso público).
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	handler := middleware.CORS(middleware.Auth(mux))

	log.Printf("Servidor escuchando en %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
