package main

import (
	"log"
	"net/http"
	"os"
	"auth-microservice/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	//Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	r := chi.NewRouter()

	//Register new routes
	r.Post("/signup", handlers.SignupHandler)
	r.Post("/verify", handlers.VerifyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s...", port)

	//Start the http server to listen
	err := http.ListenAndServe(":"+port, r); 
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
