package config

import (
	"path/filepath"
	"testing"
)

func TestConfigPath_DefaultsToHomeConfig(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	got, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	want := filepath.Join("/home/test", ".config", "repokit", "config.yaml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
