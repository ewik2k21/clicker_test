package config

import (
	"fmt"
	"os"
)

type Config struct {
	PostgresConnectionStr string
	HttpPort              string
}

type PostgresParams struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

func InitConfig() *Config {
	params := &PostgresParams{
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
	}

	if params.PostgresHost == "" {
		params.PostgresHost = "postgres"
	}

	if params.PostgresPort == "" {
		params.PostgresPort = "5432"
	}

	if params.PostgresUser == "" {
		params.PostgresUser = "user"
	}

	if params.PostgresPassword == "" {
		params.PostgresPassword = "password"
	}
	if params.PostgresDB == "" {
		params.PostgresDB = "clicks"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		params.PostgresUser,
		params.PostgresPassword,
		params.PostgresHost,
		params.PostgresPort,
		params.PostgresDB)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = ":8080"
	}

	cfg := &Config{
		PostgresConnectionStr: connStr,
		HttpPort:              httpPort,
	}

	return cfg
}
