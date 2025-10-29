package main

import (
	"log"

	"github.com/cymoo/mote/internal/app"
	"github.com/cymoo/mote/internal/config"
)

func main() {
	cfg := config.Load()
	application := app.New(cfg)

	if err := application.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
