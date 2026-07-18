package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"district-friends/internal/api"
	"district-friends/internal/db"
	"district-friends/internal/ws"
)

// @title Friends District API
// @version 1.0
// @description This is a friends feature extension backend.
// @host localhost:8080
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	dbConn := db.InitDB()

	hub := ws.NewHub()
	go hub.Run()

	r := api.SetupRouter(dbConn, hub)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
