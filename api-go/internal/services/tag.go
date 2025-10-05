package services

import "github.com/jmoiron/sqlx"

type TagService struct {
	db *sqlx.DB
}

func NewTagService(db *sqlx.DB) *TagService {
	return &TagService{db: db}
}

func (s *TagService) U(id int64) {
	panic("not implemented")
}
