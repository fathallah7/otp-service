package otp

import (
	"fmt"
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

func (h *Handler) RequestOTP(c *gin.Context) error {
	var req RequestOTPRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	if err := h.service.RequestOTP(
		c.Request.Context(),
		req.Recipient,
	); err != nil {
		return err
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "OTP request accepted",
	})

	return nil

}
