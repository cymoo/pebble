package services

import "os"

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// IsValidToken checks if the provided token matches the password set in the environment variable
func (s *AuthService) IsValidToken(token string) bool {
	password := os.Getenv("MOTE_PASSWORD")
	if password == "" {
		panic("password is not set")
	}
	return token == password
}
