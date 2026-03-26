package routes

import (
	"github.com/gin-gonic/gin"
	"fleexa/internal/api/handlers"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/health", handlers.HealthCheck)
}