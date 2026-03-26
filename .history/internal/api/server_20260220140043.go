package api

import (
	"github.com/gin-gonic/gin"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"

	"github.com/Fleexa-Graduation-Project/Backend/internal/api/routes"
)

type Server struct {
	router      *gin.Engine
	ginLambda   *ginadapter.GinLambda
}

func NewServer() *Server {
	router := gin.Default()

	routes.RegisterRoutes(router)

	return &Server{
		router:    router,
		ginLambda: ginadapter.New(router),
	}
}

func (s *Server) LambdaHandler() *ginadapter.GinLambda {
	return s.ginLambda
}