package config

import (
	"os"
	"path"
)

func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	return path.Join(home, ".config/what-it-do")
}
