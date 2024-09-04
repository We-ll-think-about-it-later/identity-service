package main

import (
	"log"

	"github.com/We-ll-think-about-it-later/identity-service/config"
	"github.com/We-ll-think-about-it-later/identity-service/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
