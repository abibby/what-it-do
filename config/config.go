package config

import (
	"os"
	"path"
)

func Dir(parts ...string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	return path.Join(append([]string{home, ".config/what-it-do"}, parts...)...)
}
