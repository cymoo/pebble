package main

import (
	"log"

	"github.com/cymoo/pebble/internal/app"
	"github.com/cymoo/pebble/internal/config"
)

func main() {
	cfg := config.Load()
	application := app.New(cfg)

	if err := application.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
