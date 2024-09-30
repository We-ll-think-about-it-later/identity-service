package http

import (
	"errors"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	_ "github.com/We-ll-think-about-it-later/identity-service/docs"

	"github.com/We-ll-think-about-it-later/identity-service/internal/controller/http/middleware"
	"github.com/We-ll-think-about-it-later/identity-service/internal/controller/http/types"
	"github.com/We-ll-think-about-it-later/identity-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Logger  *logrus.Logger
	router  *gin.Engine
	Usecase service.UserService
}

func NewServer(uc service.UserService, logger *logrus.Logger) *Server {
	logger = logger.WithField("prefix", "controller - http").Logger

	router := gin.New()
	router.Use(middleware.LoggingMiddleware(logger), gin.Recovery())

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
	// Swagger
	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	s.router.GET("/swagger/*any", swaggerHandler)

	// K8s probe
	s.router.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	auth := s.router.Group("/auth")
	{
		auth.POST("/authenticate", s.Authenticate)
		auth.POST("/token", s.GetTokens)
		auth.POST("/token/refresh", s.Refresh)
	}

	users := s.router.Group("/users")
	{
		users.POST("/:user_id/profile", s.CreateUserProfile)
		users.PATCH("/:user_id/profile", s.UpdateUserProfile)
		users.GET("/:user_id/profile", s.GetUserProfile)
	}

	s.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, types.NewErrorResponseBody(errors.New("not found")))
	})

	s.router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, types.NewErrorResponseBody(errors.New("method not allowed")))
	})
}
