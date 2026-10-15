package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() (*gin.Engine, error) {
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

	return router, nil
}
