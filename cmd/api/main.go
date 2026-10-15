package main

import (
	"log"
	"net/http"
	"otp-service/internal/config"
	"otp-service/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on the environment
	if cfg.Enviroment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create http router
	router, err := server.NewRouter()
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start the server
	log.Printf("Starting server on port %s in %s mode", cfg.Port, cfg.Enviroment)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
