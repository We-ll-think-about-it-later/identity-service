package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/We-ll-think-about-it-later/identity-service/config"
	http2 "github.com/We-ll-think-about-it-later/identity-service/internal/controller/http"
	"github.com/We-ll-think-about-it-later/identity-service/internal/repository"
	"github.com/We-ll-think-about-it-later/identity-service/internal/service"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/email"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Run(cfg config.Config) {
	log := logger.NewLogger(cfg.Level, os.Stdin)
	log.SetPrefix("app run ")

	// Initialize MongoDB
	mongo := initMongoDB(cfg, log)
	defer func() {
		if err := mongo.Disconnect(); err != nil {
			log.Fatal(err)
		}
	}()

	// Initialize services
	emailSender := initEmailSender(cfg, log)
	userService := initServices(cfg, mongo, emailSender, log)

	// Initialize and start the HTTP server
	startHTTPServer(cfg, userService, log)
}

// initMongoDB sets up the MongoDB connection and returns the client.
func initMongoDB(cfg config.Config, log *logger.Logger) *mongodb.Client {
	mongoCreds := options.Credential{
		Username: cfg.MongoDB.User,
		Password: cfg.MongoDB.Password,
	}

	mongo, err := mongodb.New(context.Background(), mongoCreds, cfg.MongoDB.Host, cfg.MongoDB.Port)
	if err != nil {
		log.Fatal(err)
	}

	return mongo
}

// initEmailSender sets up the email sender and handles errors.
func initEmailSender(cfg config.Config, log *logger.Logger) *email.EmailSender {
	login, err := email.NewEmail(cfg.Email.Login)
	if err != nil {
		log.Fatal(err)
	}

	emailSender, err := email.NewEmailSender(
		login,
		cfg.Email.Password,
		cfg.Email.SmtpHost,
		cfg.Email.SmtpPort,
	)
	if err != nil {
		log.Fatal(err)
	}

	return &emailSender
}

// initServices initializes the user and token services.
func initServices(cfg config.Config, mongo *mongodb.Client, emailSender *email.EmailSender, log *logger.Logger) service.UserService {
	dbName := "identity"

	// Initialize repositories
	userRepo := repository.NewUserRepository(mongo, dbName, "users", log)
	tokenRepo := repository.NewTokenRepository(mongo, dbName, "tokens", log)
	codeRepo := repository.NewCodeRepository(mongo, dbName, "codes", log)

	// Initialize services
	tokenService := service.NewTokenService(tokenRepo, cfg.AccessToken.LifeTime, []byte(cfg.AccessToken.Secret), log)
	userService := service.NewUserService(tokenService, emailSender, userRepo, codeRepo, log)

	return userService
}

// startHTTPServer configures and starts the HTTP server with graceful shutdown.
func startHTTPServer(cfg config.Config, userService service.UserService, log *logger.Logger) {
	// Initialize Gin router and HTTP server
	server := http2.NewServer(userService, log)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: server, // Using the Server as the HTTP handler
	}

	// Graceful shutdown setup
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Infof("Server started on 0.0.0.0:%d\n", cfg.HTTP.Port)
	<-interrupt
	log.Info("Server stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Info("Server exited properly")
}
