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

func (h *Handler) VerifyOTP(c *gin.Context) error {
	var req VerifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	valid, err := h.service.VerifyOTP(
		c.Request.Context(),
		req.Recipient,
		req.Code,
	)

	if err != nil {
		return err
	}

	if !valid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "invalid or expired OTP",
		})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully",
	})

	return nil
}
