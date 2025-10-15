package main

import (
	"log"

	"github.com/cymoo/pebble/internal/app"
	"github.com/cymoo/pebble/internal/config"
)

func main() {
	cfg := config.Load()
	application := app.New(cfg)

	if err := application.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
