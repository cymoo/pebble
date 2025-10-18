package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cymoo/pebble/internal/models"
	"github.com/jmoiron/sqlx"
)

var (
	ErrTagNotFound = errors.New("tag not found")
)

type TagService struct {
	db *sqlx.DB
}

func NewTagService(db *sqlx.DB) *TagService {
	return &TagService{db: db}
}

// GetCount returns the total count of tags
func (s *TagService) GetCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM tags`

	var count int64
	err := s.db.GetContext(ctx, &count, query)
	return count, err
}

// GetAllWithPostCount retrieves all tags with their post counts
func (s *TagService) GetAllWithPostCount(ctx context.Context) ([]models.TagWithPostCount, error) {
	query := `
		SELECT t.name, t.sticky,
			(
				SELECT COUNT(DISTINCT a.post_id)
				FROM tag_post_assoc a
				WHERE a.tag_id IN (
					SELECT id
					FROM tags
					WHERE name = t.name
					   OR name LIKE t.name || '/%'
				)
			) AS post_count
		FROM tags t
	`

	tags := []models.TagWithPostCount{}
	err := s.db.SelectContext(ctx, &tags, query)
	return tags, err
}

// GetAllWithUndeletedPostCount retrieves all tags with counts of non-deleted posts
func (s *TagService) GetAllWithUndeletedPostCount(ctx context.Context) ([]models.TagWithPostCount, error) {
	query := `
		WITH tag_posts AS (
			SELECT t.name AS tag_name, p.id AS post_id
			FROM tags t
			JOIN tag_post_assoc tpa ON t.id = tpa.tag_id
			JOIN posts p ON tpa.post_id = p.id
			WHERE p.deleted_at IS NULL
		)
		SELECT t.name AS name,
			   t.sticky AS sticky,
			   COUNT(DISTINCT tp.post_id) AS post_count
		FROM tags t
		LEFT JOIN tag_posts tp ON tp.tag_name = t.name OR tp.tag_name LIKE (t.name || '/%')
		GROUP BY t.name
	`

	var tags []models.TagWithPostCount
	err := s.db.SelectContext(ctx, &tags, query)
	return tags, err
}

// GetPosts retrieves all posts associated with a tag (including subtags)
func (s *TagService) GetPosts(ctx context.Context, name string) ([]models.Post, error) {
	namePattern := name + "/%"
	query := `
		SELECT p.*
		FROM posts p
		WHERE EXISTS (
			SELECT 1
			FROM tags t
			JOIN tag_post_assoc tp ON t.id = tp.tag_id
			WHERE tp.post_id = p.id
			AND (t.name = ? OR t.name LIKE ?)
		)
		AND p.deleted_at IS NULL
	`

	var posts []models.Post
	err := s.db.SelectContext(ctx, &posts, query, name, namePattern)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].Tags = []string{}
	}

	return posts, nil
}

// InsertOrUpdate inserts a new tag or updates its sticky status
func (s *TagService) InsertOrUpdate(ctx context.Context, name string, sticky bool) error {
	now := time.Now().UnixMilli()

	query := `
		INSERT INTO tags (name, sticky, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			sticky = excluded.sticky,
			updated_at = excluded.updated_at
	`

	_, err := s.db.ExecContext(ctx, query, name, sticky, now, now)
	return err
}

// DeleteAssociatedPosts soft-deletes all posts associated with a tag
func (s *TagService) DeleteAssociatedPosts(ctx context.Context, name string) error {
	now := time.Now().UnixMilli()
	namePattern := name + "/%"

	query := `
		UPDATE posts
		SET deleted_at = ?
		WHERE id IN (
			SELECT post_id
			FROM tag_post_assoc
			WHERE tag_id IN (
				SELECT id
				FROM tags
				WHERE name = ? OR name LIKE ?
			)
		)
	`

	_, err := s.db.ExecContext(ctx, query, now, name, namePattern)
	return err
}

// RenameOrMerge renames a tag or merges it with an existing tag
func (s *TagService) RenameOrMerge(ctx context.Context, name, newName string) error {
	if name == newName {
		return nil
	}

	// Check for invalid hierarchy
	if strings.HasPrefix(newName, name+"/") {
		newDepth := strings.Count(newName, "/")
		oldDepth := strings.Count(name, "/")
		if newDepth > oldDepth {
			return fmt.Errorf("cannot move %q to a subtag of itself %q", name, newName)
		}
	}

	namePattern := name + "/%"

	// Get all affected tags
	query := `
		SELECT * FROM tags
		WHERE name = ? OR name = ? OR name LIKE ?
	`

	var affectedTags []models.Tag
	err := s.db.SelectContext(ctx, &affectedTags, query, name, newName, namePattern)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Find source and target tags
	var sourceTag *models.Tag
	var targetTag *models.Tag
	descendants := make([]*models.Tag, 0)

	for i := range affectedTags {
		tag := &affectedTags[i]
		if tag.Name == name {
			sourceTag = tag
		} else if tag.Name == newName {
			targetTag = tag
		} else {
			descendants = append(descendants, tag)
		}
	}

	// Create source tag if it doesn't exist
	if sourceTag == nil {
		newTag, err := s.create(ctx, tx, name)
		if err != nil {
			return err
		}
		sourceTag = newTag
	}

	// Sort descendants by depth (deepest first)
	sort.Slice(descendants, func(i, j int) bool {
		return strings.Count(descendants[i].Name, "/") > strings.Count(descendants[j].Name, "/")
	})

	// Process descendants
	for _, descendant := range descendants {
		newDescendantName := replacePrefix(descendant.Name, name, newName)
		targetDescendant, err := s.findByName(ctx, tx, newDescendantName)
		if err != nil {
			return err
		}

		if targetDescendant != nil {
			// Target exists - merge
			if err := s.merge(ctx, tx, descendant, targetDescendant); err != nil {
				return err
			}
		} else {
			// Target doesn't exist - rename
			if err := s.rename(ctx, tx, descendant, newDescendantName); err != nil {
				return err
			}
		}
	}

	// Process source tag
	if targetTag != nil {
		if err := s.merge(ctx, tx, sourceTag, targetTag); err != nil {
			return err
		}
	} else {
		if err := s.rename(ctx, tx, sourceTag, newName); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Helper functions

func (s *TagService) findOrCreate(ctx context.Context, tx *sqlx.Tx, name string) (*models.Tag, error) {
	tag, err := s.findByName(ctx, tx, name)
	if err != nil {
		return nil, err
	}
	if tag != nil {
		return tag, nil
	}

	return s.create(ctx, tx, name)
}

func (s *TagService) findByName(ctx context.Context, tx *sqlx.Tx, name string) (*models.Tag, error) {
	query := `SELECT * FROM tags WHERE name = ?`

	var tag models.Tag
	err := tx.GetContext(ctx, &tag, query, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func (s *TagService) create(ctx context.Context, tx *sqlx.Tx, name string) (*models.Tag, error) {
	now := time.Now().UnixMilli()

	query := `
		INSERT INTO tags (name, sticky, created_at, updated_at)
		VALUES (?, false, ?, ?)
		RETURNING id
	`

	var id int64
	err := tx.QueryRowContext(ctx, query, name, now, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &models.Tag{
		ID:        id,
		Name:      name,
		Sticky:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *TagService) rename(ctx context.Context, tx *sqlx.Tx, tag *models.Tag, newName string) error {
	now := time.Now().UnixMilli()

	// Update tag name
	query := `UPDATE tags SET name = ?, updated_at = ? WHERE id = ?`
	_, err := tx.ExecContext(ctx, query, newName, now, tag.ID)
	if err != nil {
		return err
	}

	// Update post content
	sourcePattern := fmt.Sprintf(">%s<", tag.Name)
	targetPattern := fmt.Sprintf(">%s<", newName)

	updateQuery := `
		UPDATE posts
		SET content = REPLACE(content, ?, ?)
		WHERE id IN (
			SELECT post_id
			FROM tag_post_assoc
			WHERE tag_id = ?
		)
	`

	_, err = tx.ExecContext(ctx, updateQuery, sourcePattern, targetPattern, tag.ID)
	return err
}

func (s *TagService) merge(ctx context.Context, tx *sqlx.Tx, sourceTag, targetTag *models.Tag) error {
	// Update post content
	sourcePattern := fmt.Sprintf(">%s<", sourceTag.Name)
	targetPattern := fmt.Sprintf(">%s<", targetTag.Name)

	updateQuery := `
		UPDATE posts
		SET content = REPLACE(content, ?, ?)
		WHERE id IN (
			SELECT post_id
			FROM tag_post_assoc
			WHERE tag_id = ?
		)
	`

	_, err := tx.ExecContext(ctx, updateQuery, sourcePattern, targetPattern, sourceTag.ID)
	if err != nil {
		return err
	}

	// Insert new tag associations (ignore if they already exist)
	insertQuery := `
		INSERT OR IGNORE INTO tag_post_assoc (post_id, tag_id)
		SELECT post_id, ? as tag_id
		FROM tag_post_assoc
		WHERE tag_id = ?
	`

	_, err = tx.ExecContext(ctx, insertQuery, targetTag.ID, sourceTag.ID)
	if err != nil {
		return err
	}

	// Delete old tag associations
	deleteQuery := `DELETE FROM tag_post_assoc WHERE tag_id = ?`
	_, err = tx.ExecContext(ctx, deleteQuery, sourceTag.ID)
	if err != nil {
		return err
	}

	// Optionally delete the source tag itself
	// deleteTagQuery := `DELETE FROM tags WHERE id = ?`
	// _, err = tx.ExecContext(ctx, deleteTagQuery, sourceTag.ID)

	return nil
}

// replacePrefix replaces the prefix of a string
func replacePrefix(s, oldPrefix, newPrefix string) string {
	if !strings.HasPrefix(s, oldPrefix) {
		return s
	}
	return newPrefix + s[len(oldPrefix):]
}
