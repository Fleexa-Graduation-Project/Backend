package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "online",
		"service":   "fleexa-api",
		"timestamp": time.Now().Unix(),
	})
}