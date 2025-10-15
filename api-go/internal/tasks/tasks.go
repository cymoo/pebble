package tasks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cymoo/mita"
	"github.com/jmoiron/sqlx"
)

// DeleteOldPosts deletes posts that were marked as deleted more than 30 days ago
func DeleteOldPosts(ctx context.Context) error {
	db := ctx.Value(mita.CtxtKey("db")).(*sqlx.DB)

	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30).UnixMilli()

	result, err := db.Exec("DELETE FROM posts WHERE deleted_at < $1", thirtyDaysAgo)
	if err != nil {
		return fmt.Errorf("error deleting old posts: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[Daily] successfully deleted %d posts", rowsAffected)
	}
	return nil
}
