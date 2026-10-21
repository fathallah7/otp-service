package server

import (
	"errors"
	"log"
	"net/http"

	"otp-service/internal/otp"

	"github.com/gin-gonic/gin"
)

type ErrorHandlerFunc func(*gin.Context) error

func HandleErrors(handler ErrorHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handler(c); err != nil {
			log.Printf("Request failed: %v", err)
			respondError(c, err)
		}
	}
}

func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "An unexpected error occurred"

	switch {
	case errors.Is(err, otp.ErrInvalidRequest):
		status = http.StatusBadRequest
		code = "invalid_request"
		message = "Invalid request body"

	case errors.Is(err, otp.ErrInvalidRecipient):
		status = http.StatusBadRequest
		code = "invalid_recipient"
		message = "Invalid email or phone number"

	case errors.Is(err, otp.ErrDeliveryFailed):
		status = http.StatusBadGateway
		code = "delivery_failed"
		message = "Unable to send verification code"
	}

	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
