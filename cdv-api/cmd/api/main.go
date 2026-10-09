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
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"cdv-api/internal/cabins"
	"cdv-api/internal/database"
	"cdv-api/internal/images"
	"cdv-api/internal/middleware"
	"cdv-api/internal/notifications"
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
	notificationRepository := notifications.NewRepository(pool)
	notificationService := notifications.NewService(notificationRepository)
	notificationController := notifications.NewController(notificationService)

	userRepository := users.NewRepository(pool)
	userService := users.NewService(userRepository, notificationService)
	userController := users.NewController(userService)

	cabinRepository := cabins.NewRepository(pool)
	cabinService := cabins.NewService(cabinRepository, notificationService)
	cabinController := cabins.NewController(cabinService)

	reservationRepository := reservations.NewRepository(pool)
	reservationService := reservations.NewService(reservationRepository, notificationService)
	reservationController := reservations.NewController(reservationService)

	imageRepository := images.NewRepository(pool)
	imageService := images.NewService(imageRepository)
	imageController := images.NewController(imageService)

	// -------------------------
	// Jobs en segundo plano
	// -------------------------
	// Finaliza las reservaciones cuya fecha de fin ya pasó y notifica al
	// huésped, al anfitrión y a los administradores. Se ejecuta al arrancar y
	// luego cada 6 horas; los errores nunca detienen la API.
	go func() {
		const finalizeInterval = 6 * time.Hour

		finalize := func() {
			jobCtx := context.Background()
			count, err := reservationService.FinalizePastReservations(jobCtx)
			if err != nil {
				log.Printf("reservations: error al finalizar reservaciones vencidas: %v", err)
				return
			}
			if count > 0 {
				log.Printf("reservations: %d reservación(es) finalizada(s)", count)
			}
		}

		finalize()

		ticker := time.NewTicker(finalizeInterval)
		defer ticker.Stop()
		for range ticker.C {
			finalize()
		}
	}()

	// -------------------------
	// Router
	// -------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cdv-api/users", userController.GetUsers)
	mux.HandleFunc("POST /cdv-api/login", userController.LoginUser)
	mux.HandleFunc("POST /cdv-api/users", userController.RegisterUser)
	mux.HandleFunc("GET /cdv-api/users/{id}", userController.GetUserByID)
	mux.HandleFunc("PUT /cdv-api/users/{id}", userController.UpdateUser)
	mux.HandleFunc("PATCH /cdv-api/users/{id}/role", userController.UpdateUserRole)

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
	mux.HandleFunc("GET /cdv-api/reservations/user/{userId}", reservationController.GetReservationsByUser)
	mux.HandleFunc("GET /cdv-api/reservations/cabin/{cabinId}", reservationController.GetReservationsByCabin)
	mux.HandleFunc("GET /cdv-api/reservations/{id}", reservationController.GetReservationByID)

	// Buzón de notificaciones (personal: siempre el usuario del token).
	mux.HandleFunc("GET /cdv-api/notifications", notificationController.GetNotifications)
	mux.HandleFunc("GET /cdv-api/notifications/unread-count", notificationController.GetUnreadCount)
	mux.HandleFunc("POST /cdv-api/notifications/read-all", notificationController.MarkAllRead)
	mux.HandleFunc("PATCH /cdv-api/notifications/{id}/read", notificationController.MarkRead)
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
