package config

import (
	"time"
)

type Config struct {
	HTTP  HTTPConfig
	DB    DBConfig
	Redis RedisConfig
}

type HTTPConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DBConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Host:         "localhost",
			Port:         "8080",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
		DB: DBConfig{
			DSN:          "file:app.db?cache=shared&_fk=true&_journal_mode=WAL",
			MaxOpenConns: 25,
			MaxIdleConns: 25,
			MaxIdleTime:  15 * time.Minute,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}
}
