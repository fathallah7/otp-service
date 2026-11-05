package bootstrap

import (
	"fmt"
	"log"
	"net/http"

	"otp-service/internal/otp"
	"otp-service/internal/server"

	"github.com/gin-gonic/gin"
)

// RunAPI initializes the HTTP API and serves requests until the server stops.
func RunAPI() error {
	common, err := NewCommon()
	if err != nil {
		return err
	}
	defer common.Close()

	// Set Gin mode based on the environment
	if common.Config.Enviroment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Handler and inject Service
	otpHandler := otp.NewHandler(common.OTPService)

	// Create http router
	router, err := server.NewRouter(otpHandler)
	if err != nil {
		return fmt.Errorf("create router: %w", err)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + common.Config.Port,
		Handler: router,
	}

	// Start the server
	log.Printf(
		"Starting server on port %s in %s mode",
		common.Config.Port,
		common.Config.Enviroment,
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}
