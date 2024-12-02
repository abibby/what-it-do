package config

import (
	"errors"
	"os"
	"path"

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

var Cfg *Config

func Load() *Config {
	err := godotenv.Load("./.env", path.Join(Dir(), "env"))
	if errors.Is(err, os.ErrNotExist) {
		// fall through
	} else if err != nil {
		panic(err)
	}
	err = godotenv.Load(path.Join(Dir(), "env"))
	if errors.Is(err, os.ErrNotExist) {
		// fall through
	} else if err != nil {
		panic(err)
	}

	Cfg = &Config{
		Jira: &JiraConfig{
			Username: env.String("JIRA_USERNAME", ""),
			Password: env.String("JIRA_PASSWORD", ""),
			BaseURL:  env.String("JIRA_BASE_URL", ""),
		},
	}
	return Cfg
}

func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	return path.Join(home, ".config/what-it-do")
}
