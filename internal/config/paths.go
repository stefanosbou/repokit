package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigPath returns the absolute path to the repokit config file,
// defaults to ~/.config/repokit/config.yaml.
func ConfigPath() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		home = h
	}
	return filepath.Join(home, ".config", "repokit", "config.yaml"), nil
}
