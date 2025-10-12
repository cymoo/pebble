package services

import "os"

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) IsValidToken(token string) bool {
	password := os.Getenv("PEBBLE_PASSWORD")
	if password == "" {
		panic("password is not set")
	}
	return token == password
}
