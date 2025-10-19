package main

import (
	"time"

	"github.com/cymoo/pebble/internal/app"
	"github.com/cymoo/pebble/internal/config"
)

func main() {
	expirePost(1)
}

// expirePost sets the deleted_at field of a post to 31 days ago, effectively expiring it
// postID: ID of the post to expire
func expirePost(postID int64) {
	cfg := config.Load()

	application := app.New(cfg)

	db := application.GetDB()

	thirtyOneDaysAgo := time.Now().UTC().AddDate(0, 0, -31).UnixMilli()

	_, err := db.Exec(
		"UPDATE posts SET deleted_at = $1 WHERE id = $2",
		thirtyOneDaysAgo,
		postID,
	)
	if err != nil {
		panic(err)
	}
}
