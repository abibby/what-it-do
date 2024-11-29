package config

import (
	"errors"
	"os"

	"github.com/abibby/salusa/env"
	"github.com/joho/godotenv"
)

type Config struct {
	Jira *JiraConfig
}

type JiraConfig struct {
	Username string
	Password string
	BaseURL  string
}

func Load() *Config {
	err := godotenv.Load("./.env")
	if errors.Is(err, os.ErrNotExist) {
		// fall through
	} else if err != nil {
		panic(err)
	}

	return &Config{
		Jira: &JiraConfig{
			Username: env.String("JIRA_USERNAME", ""),
			Password: env.String("JIRA_PASSWORD", ""),
			BaseURL:  env.String("JIRA_BASE_URL", ""),
		},
	}
}
