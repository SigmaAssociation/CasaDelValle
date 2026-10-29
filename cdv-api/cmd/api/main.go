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
	"cdv-api/internal/payments"
	"cdv-api/internal/refunds"
	reportincome "cdv-api/internal/reports/reportIncome"
	reportreservations "cdv-api/internal/reports/reportReservations"
	"cdv-api/internal/reservations"
	"cdv-api/internal/seasonprices"
	"cdv-api/internal/seasons"
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

	refundRepository := refunds.NewRepository(pool)
	refundService := refunds.NewService(refundRepository)
	refundController := refunds.NewController(refundService, reservationService)
	reservationService.SetRefundRecorder(refundService)

	seasonRepository := seasons.NewRepository(pool)
	seasonService := seasons.NewService(seasonRepository)
	seasonController := seasons.NewController(seasonService)

	imageRepository := images.NewRepository(pool)
	imageService := images.NewService(imageRepository)
	imageController := images.NewController(imageService)

	seasonPriceRepository := seasonprices.NewRepository(pool)
	seasonPriceService := seasonprices.NewService(seasonPriceRepository)
	seasonPriceController := seasonprices.NewController(seasonPriceService)

	reportRepository := reportreservations.NewRepository(pool)
	reportService := reportreservations.NewService(reportRepository)
	reportController := reportreservations.NewController(reportService)

	incomeReportRepository := reportincome.NewRepository(pool)
	incomeReportService := reportincome.NewService(incomeReportRepository)
	incomeReportController := reportincome.NewController(incomeReportService)
	paymentRepository := payments.NewRepository(pool)
	paymentService := payments.NewService(paymentRepository, notificationService)
	paymentController := payments.NewController(paymentService)

	go func() {
		const finalizeInterval = 6 * time.Hour

		finalize := func() {
			jobCtx := context.Background()
			count, err := reservationService.AdvanceReservationStates(jobCtx)
			if err != nil {
				log.Printf("reservations: error al avanzar los estados de las reservaciones: %v", err)
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

	mux.HandleFunc("GET /cdv-api/cabins/availability/{cabinId}", reservationController.GetCabinAvailability)

	mux.HandleFunc("GET /cdv-api/seasons", seasonController.ListSeasons)
	mux.HandleFunc("POST /cdv-api/seasons", seasonController.CreateSeason)
	mux.HandleFunc("PUT /cdv-api/seasons/{id}", seasonController.UpdateSeason)
	mux.HandleFunc("DELETE /cdv-api/seasons/{id}", seasonController.DeleteSeason)
	mux.HandleFunc("GET /cdv-api/cabin-season-prices/cabin/{cabinId}", seasonController.ListCabinSeasonPrices)
	mux.HandleFunc("POST /cdv-api/cabin-season-prices/cabin/{cabinId}", seasonController.SetCabinSeasonPrice)
	mux.HandleFunc("DELETE /cdv-api/cabin-season-prices/{id}", seasonController.DeleteCabinSeasonPrice)
	mux.HandleFunc("POST /cdv-api/seasons/quote", seasonController.GetPriceQuote)

	mux.HandleFunc("GET /cdv-api/cabins/{id}", cabinController.GetCabinByID)
	mux.HandleFunc("PUT /cdv-api/cabins/{id}", cabinController.UpdateCabin)
	mux.HandleFunc("DELETE /cdv-api/cabins/{id}", cabinController.DeleteCabin)

	mux.HandleFunc("POST /cdv-api/cabins/images/{cabinId}", imageController.CreateImage)
	mux.HandleFunc("GET /cdv-api/cabins/images/{cabinId}", imageController.GetImagesByCabin)
	mux.HandleFunc("GET /cdv-api/cabins/images/{cabinId}/principal", imageController.GetMainImageByCabin)
	mux.HandleFunc("GET /cdv-api/images/{id}", imageController.GetImageByID)
	mux.HandleFunc("DELETE /cdv-api/images/{id}", imageController.DeleteImage)

	mux.HandleFunc("DELETE /cdv-api/cabins/{cabinId}/season-prices/{id}", seasonPriceController.DeleteSeasonPrice)

	mux.HandleFunc("POST /cdv-api/reservations", reservationController.CreateReservation)
	mux.HandleFunc("PUT /cdv-api/reservations/{id}", reservationController.UpdateReservation)
	mux.HandleFunc("PATCH /cdv-api/reservations/{id}/cancel", reservationController.CancelReservation)
	mux.HandleFunc("PATCH /cdv-api/reservations/{id}/checkin", reservationController.Checkin)
	mux.HandleFunc("PATCH /cdv-api/reservations/{id}/checkout", reservationController.Checkout)
	mux.HandleFunc("POST /cdv-api/reservations/{id}/payment", reservationController.PayAnticipo)
	mux.HandleFunc("GET /cdv-api/payments/reservation/{reservationId}", paymentController.GetReservationPayments)
	mux.HandleFunc("GET /cdv-api/reservations/cabin/{cabinId}", reservationController.GetReservationsByCabin)
	mux.HandleFunc("GET /cdv-api/reservations/user/{userId}", reservationController.GetReservationsByUser)
	mux.HandleFunc("GET /cdv-api/reservations/{id}", reservationController.GetReservationByID)
	mux.HandleFunc("GET /cdv-api/reservations/balance/{id}", reservationController.GetReservationBalance)

	mux.HandleFunc("POST /cdv-api/refunds/reservation/{id}", refundController.CreateRefund)
	mux.HandleFunc("GET /cdv-api/refunds/reservation/{id}", refundController.GetRefunds)

	mux.HandleFunc("GET /cdv-api/notifications", notificationController.GetNotifications)
	mux.HandleFunc("GET /cdv-api/notifications/unread-count", notificationController.GetUnreadCount)
	mux.HandleFunc("POST /cdv-api/notifications/read-all", notificationController.MarkAllRead)
	mux.HandleFunc("PATCH /cdv-api/notifications/{id}/read", notificationController.MarkRead)

	//Reportes
	mux.HandleFunc("GET /cdv-api/reports/reservations", reportController.GetReservationsReport)
	mux.HandleFunc("GET /cdv-api/reports/income", incomeReportController.GetIncomeReport)

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

// TEMPORAL
