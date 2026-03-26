package api

import (
	"github.com/gin-gonic/gin"
	"github.com/Fleexa-Graduation-Project/Backend/internal/api/routes"
)

type Server struct {
	router *gin.Engine
}

func NewServer() *Server {
	router := gin.Default()

	routes.RegisterRoutes(router)

	return &Server{
		router: router,
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}