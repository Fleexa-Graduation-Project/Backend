package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Fleexa-Graduation-Project/Backend/internal/api/handlers"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/health", handlers.HealthCheck)
}