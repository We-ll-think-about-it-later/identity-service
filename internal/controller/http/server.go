package http

import (
	"errors"
	"net/http"

	"github.com/We-ll-think-about-it-later/identity-service/internal/controller/http/middleware"
	"github.com/We-ll-think-about-it-later/identity-service/internal/controller/http/types"
	"github.com/We-ll-think-about-it-later/identity-service/internal/service"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Logger  *logger.Logger
	router  *gin.Engine
	Usecase service.UserService
}

func NewServer(uc service.UserService, logger *logger.Logger) *Server {
	logger.SetPrefix("controller - http ")

	router := gin.Default()
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(gin.Recovery())

	s := Server{
		router:  router,
		Logger:  logger,
		Usecase: uc,
	}

	s.configureRouter()

	return &s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) configureRouter() {
	auth := s.router.Group("/auth")
	{
		auth.POST("/signup", s.Signup)
		auth.POST("/login", s.Login)
		auth.POST("/get_tokens", s.GetTokens)
		auth.POST("/refresh", s.Refresh)
	}

	user := s.router.Group("/users")
	user.POST("/me", s.UserMe)

	s.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, types.NewErrorResponseBody(errors.New("not found")))
	})

	s.router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, types.NewErrorResponseBody(errors.New("method not allowed")))
	})
}
