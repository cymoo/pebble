package models

import (
	"database/sql"
)

// Post represents a post entity
type Post struct {
	ID      int64  `json:"id" db:"id"`
	Content string `json:"content" db:"content"`
	// Files         json.RawMessage `json:"files,omitempty" db:"files"`
	Files         string         `json:"files,omitempty" db:"files"`
	Color         sql.NullString `json:"color,omitempty" db:"color"`
	Shared        bool           `json:"shared" db:"shared"`
	DeletedAt     sql.NullInt64  `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedAt     int64          `json:"created_at" db:"created_at"`
	UpdatedAt     int64          `json:"updated_at" db:"updated_at"`
	ParentID      sql.NullInt64  `json:"-" db:"parent_id"`
	ChildrenCount int64          `json:"children_count" db:"children_count"`

	// Additional fields not in DB
	Parent *Post    `json:"parent,omitempty"`
	Score  *float64 `json:"score,omitempty"`
	Tags   []string `json:"tags"`
}

// FileInfo represents file metadata
type FileInfo struct {
	URL      string  `json:"url"`
	ThumbURL *string `json:"thumb_url,omitempty"`
	Size     *uint64 `json:"size,omitempty"`
	Width    *uint32 `json:"width,omitempty"`
	Height   *uint32 `json:"height,omitempty"`
}

// Tag represents a tag entity
type Tag struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Sticky    bool   `json:"sticky" db:"sticky"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
	UpdatedAt int64  `json:"updated_at" db:"updated_at"`
}

// TagWithPostCount represents a tag with its post count
type TagWithPostCount struct {
	Name      string `json:"name" db:"name"`
	Sticky    bool   `json:"sticky" db:"sticky"`
	PostCount int64  `json:"post_count" db:"post_count"`
}

// CreatePostRequest represents the request to create a post
type CreatePostRequest struct {
	Content  string     `json:"content"`
	Files    []FileInfo `json:"files,omitempty"`
	Color    *string    `json:"color,omitempty"`
	Shared   *bool      `json:"shared,omitempty"`
	ParentID *int64     `json:"parent_id,omitempty"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	ID       int64       `json:"id"`
	Content  *string     `json:"content,omitempty"`
	Shared   *bool       `json:"shared,omitempty"`
	Files    *[]FileInfo `json:"files,omitempty"`
	Color    *string     `json:"color,omitempty"`
	ParentID *int64      `json:"parent_id,omitempty"`
}

// PostFilterOptions represents filtering options for posts
type PostFilterOptions struct {
	Cursor    *int64  `json:"cursor,omitempty"`
	Deleted   bool    `json:"deleted"`
	ParentID  *int64  `json:"parent_id,omitempty"`
	Color     *string `json:"color,omitempty"`
	Tag       *string `json:"tag,omitempty"`
	Shared    *bool   `json:"shared,omitempty"`
	HasFiles  *bool   `json:"has_files,omitempty"`
	OrderBy   string  `json:"order_by"`
	Ascending bool    `json:"ascending"`
	StartDate *int64  `json:"start_date,omitempty"`
	EndDate   *int64  `json:"end_date,omitempty"`
}

// PostPagination represents paginated posts
type PostPagination struct {
	Posts  []Post `json:"posts"`
	Cursor int64  `json:"cursor"`
	Size   int64  `json:"size"`
}

// PostStats represents statistics about posts
type PostStats struct {
	PostCount int64 `json:"post_count"`
	TagCount  int64 `json:"tag_count"`
	DayCount  int64 `json:"day_count"`
}

// CreateResponse represents the response after creating a post
type CreateResponse struct {
	ID        int64 `json:"id"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

type Id struct {
	Id int64 `json:"id"`
}
