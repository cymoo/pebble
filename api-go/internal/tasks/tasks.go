package tasks

import (
	"context"
	"log"
)

func Cleanup(ctx context.Context) error {
	log.Printf("start to clean up deleted expired posts...")
	return nil
}
