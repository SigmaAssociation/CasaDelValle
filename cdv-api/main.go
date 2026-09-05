package main

import (
	"cdv-api/controllers"
	"cdv-api/repositories"
	"cdv-api/services"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < 5; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Error al conectar a la base de datos, reintentando en 3 segundos... (intento %d/5)\n", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("Error al conectar a la base de datos:", err)
	}
	log.Println("Conexión a PostgreSQL establecida")

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	http.HandleFunc("/cdv-api/users", userController.GetUsers)
	http.HandleFunc("/cdv-api/register", userController.RegisterUser)

	log.Println("Servidor iniciado en http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
