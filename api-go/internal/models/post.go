package models

import (
	"database/sql"
	"time"
)

type Post struct {
	ID            int64          `db:"id" json:"id"`
	Content       string         `db:"content" json:"content"`
	Files         sql.NullString `db:"files" json:"files,omitempty"`
	Color         sql.NullString `db:"color" json:"color,omitempty"`
	Shared        bool           `db:"shared" json:"shared"`
	DeletedAt     sql.NullInt64  `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedAt     int64          `db:"created_at" json:"created_at"`
	UpdatedAt     int64          `db:"updated_at" json:"updated_at"`
	ParentID      sql.NullInt64  `db:"parent_id" json:"parent_id,omitempty"`
	ChildrenCount int            `db:"children_count" json:"children_count"`
}

type CreatePostRequest struct {
	Content  string  `json:"content"`
	Files    *string `json:"files,omitempty"`
	Color    *string `json:"color,omitempty"`
	Shared   bool    `json:"shared"`
	ParentID *int64  `json:"parent_id,omitempty"`
}

type UpdatePostRequest struct {
	Content *string `json:"content,omitempty"`
	Files   *string `json:"files,omitempty"`
	Color   *string `json:"color,omitempty"`
	Shared  *bool   `json:"shared,omitempty"`
}

func (p *Post) BeforeCreate() {
	now := time.Now().UnixNano()
	p.CreatedAt = now
	p.UpdatedAt = now
}

func (p *Post) BeforeUpdate() {
	p.UpdatedAt = time.Now().UnixNano()
}
