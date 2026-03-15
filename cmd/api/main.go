package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/relchzen/cafe-blog/internal/database"
	"github.com/relchzen/cafe-blog/internal/handlers"
	"github.com/relchzen/cafe-blog/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err := database.CreateTables(); err != nil {
		log.Fatal("Failed to create tables: ", err)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/api/login", handlers.Login)
	http.HandleFunc("/api/register", handlers.Register)

	http.HandleFunc("/api/cafes", middleware.AuthMiddleware(handlers.CreateCafe))

	log.Printf("Server starting at port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status": "ok",
	}

	json.NewEncoder(w).Encode(response)
}
