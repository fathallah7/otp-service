package server

import (
	"net/http"
	"otp-service/internal/otp"

	"github.com/gin-gonic/gin"
)

func NewRouter(otpHandler *otp.Handler) (*gin.Engine, error) {
	router := gin.Default()

	// Set trusted proxies to nil to disable proxy checking
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes
	api := router.Group("/api/v1")

	// OTP routes
	otpRoutes := api.Group("/otp")
	otpRoutes.POST("/request", otpHandler.RequestOTP)

	return router, nil
}
