package otp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RequestOTP(c *gin.Context) {
	var req RequestOTPRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := h.service.RequestOTP(c.Request.Context(), req.Recipient)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRecipient):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid email or phone number",
			})

		case errors.Is(err, ErrNotImplemented):
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "OTP delivery not implemented",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "OTP request accepted",
	})

}
