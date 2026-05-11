package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  string      `yaml:"version"`
	Repos    []RepoEntry `yaml:"repos"`
	Settings Settings    `yaml:"settings"`
}

type RepoEntry struct {
	Path string `yaml:"path" json:"path"`
	Name string `yaml:"name" json:"name"`
}

type Settings struct {
	Parallel     int           `yaml:"parallel"`
	PullStrategy string        `yaml:"pull_strategy"`
	Clean        CleanSettings `yaml:"clean"`
}

type CleanSettings struct {
	StaleAfterDays    int      `yaml:"stale_after_days"`
	ProtectedBranches []string `yaml:"protected_branches"`
}

// Default returns a fully populated Config with the documented v1 defaults.
func Default() *Config {
	return &Config{
		Version: "1",
		Repos:   []RepoEntry{},
	}
}

// Load reads a config from disk. When path is "", the default is used.
// A missing file returns Default(), nil.
func Load(path string) (*Config, error) {
	if path == "" {
		p, err := ConfigPath()
		if err != nil {
			return nil, err
		}
		path = p
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(path string, cfg *Config) error {
	if path == "" {
		p, err := ConfigPath()
		if err != nil {
			return err
		}
		path = p
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}
