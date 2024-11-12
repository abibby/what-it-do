package config

import (
	"errors"
	"os"

	"github.com/abibby/salusa/database"
	"github.com/abibby/salusa/database/dialects/sqlite"
	"github.com/abibby/salusa/env"
	"github.com/abibby/salusa/event"
	"github.com/abibby/what-it-do/services/sms"
	"github.com/joho/godotenv"
)

type Config struct {
	Port     int
	BasePath string

	Database database.Config
	SMS      sms.Config
	Queue    event.Config
}

func Load() *Config {
	err := godotenv.Load("./.env")
	if errors.Is(err, os.ErrNotExist) {
		// fall through
	} else if err != nil {
		panic(err)
	}

	return &Config{
		Port:     env.Int("PORT", 2303),
		BasePath: env.String("BASE_PATH", ""),
		Database: sqlite.NewConfig(env.String("DATABASE_PATH", "./db.sqlite")),
		Queue:    event.NewChannelQueueConfig(),
	}
}

func (c *Config) GetHTTPPort() int {
	return c.Port
}
func (c *Config) GetBaseURL() string {
	return c.BasePath
}

func (c *Config) DBConfig() database.Config {
	return c.Database
}
func (c *Config) SMSConfig() sms.Config {
	return c.SMS
}
func (c *Config) QueueConfig() event.Config {
	return c.Queue
}
