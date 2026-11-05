package main

import (
	"log"
	"net/http"

	"otp-service/internal/bootstrap"
	"otp-service/internal/otp"
	"otp-service/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize shared dependencies
	deps, err := bootstrap.NewCommon()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer deps.Close()

	// Set Gin mode based on the environment
	if deps.Config.Enviroment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Handler and inject Service
	otpHandler := otp.NewHandler(deps.OTPService)

	// Create http router
	router, err := server.NewRouter(otpHandler)
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + deps.Config.Port,
		Handler: router,
	}

	// Start the server
	log.Printf("Starting server on port %s in %s mode", deps.Config.Port, deps.Config.Enviroment)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
