package config

import (
	"errors"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		HTTP
		Log
		AccessToken
		MongoDB
		Email
	}

	// HTTP -.
	HTTP struct {
		Port int `env-required:"true" env:"HTTP_PORT" env-default:"8080"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" env:"LOG_LEVEL"`
	}

	// AccessToken -.
	AccessToken struct {
		LifeTime int64  `env-required:"true" env:"ACCESS_TOKEN_LIFETIME"`
		Secret   string `env-required:"true" env:"ACCESS_TOKEN_SECRET"`
	}

	// MongoDB -.
	MongoDB struct {
		Host     string `env-required:"true" env:"MONGO_HOST" env-default:"localhost"`
		Port     int    `env-required:"true" env:"MONGO_PORT" env-default:"27017"`
		User     string `env-required:"true" env:"MONGO_INITDB_ROOT_USERNAME"`
		Password string `env-required:"true" env:"MONGO_INITDB_ROOT_PASSWORD"`
	}

	// Email -.
	Email struct {
		Login    string `env-required:"true" env:"EMAIL_LOGIN"`
		Password string `env-required:"true" env:"EMAIL_PASSWORD"`
		SmtpHost string `env-required:"true" env:"EMAIL_SMTP_HOST"`
		SmtpPort int    `env-required:"true" env:"EMAIL_SMTP_PORT"`
	}
)

// NewConfig returns app config.
func NewConfig() (Config, error) {
	var errFile error

	cfg := &Config{}
	errEnv := cleanenv.ReadEnv(cfg)

	// fallback to .env file
	if errEnv != nil {
		errFile = cleanenv.ReadConfig(".env", cfg)
	}

	if errFile != nil {
		return *cfg, errors.Join(errEnv, errFile)
	}

	return *cfg, nil
}
