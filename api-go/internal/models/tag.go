package models

import "time"

type Tag struct {
	ID        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Sticky    bool   `db:"sticky" json:"sticky"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
	UpdatedAt int64  `db:"updated_at" json:"updated_at"`
}

type CreateTagRequest struct {
	Name   string `json:"name"`
	Sticky bool   `json:"sticky"`
}

func (t *Tag) BeforeCreate() {
	now := time.Now().UnixNano()
	t.CreatedAt = now
	t.UpdatedAt = now
}
