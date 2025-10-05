package services

import "github.com/jmoiron/sqlx"

type PostService struct {
	db *sqlx.DB
}

func NewPostService(db *sqlx.DB) *PostService {
	return &PostService{db: db}
}

func (s *PostService) GetById(id int64) {
	panic("not implemented")
}
