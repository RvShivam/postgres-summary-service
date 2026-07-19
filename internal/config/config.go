package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	External ExternalConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ExternalConfig struct {
	BaseURL string
}

func getRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s environment variable is required", key)
	}

	return value, nil
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := getRequiredEnv("PORT")
	if err != nil {
		return nil, err
	}

	dbHost, err := getRequiredEnv("DB_HOST")
	if err != nil {
		return nil, err
	}

	dbPort, err := getRequiredEnv("DB_PORT")
	if err != nil {
		return nil, err
	}

	dbUser, err := getRequiredEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := getRequiredEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbName, err := getRequiredEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	externalURL, err := getRequiredEnv("EXTERNAL_SERVICE_URL")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
	Server: ServerConfig{
		Port: port,
	},
	Database: DatabaseConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Name:     dbName,
	},
	External: ExternalConfig{
		BaseURL: externalURL,
	},
}
	return cfg, nil
}