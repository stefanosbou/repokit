package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingFile_ReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Version != "1" {
		t.Errorf("default version = %q, want %q", cfg.Version, "1")
	}
	if cfg.Settings.PullStrategy != "ff-only" {
		t.Errorf("default pull strategy = %q, want %q", cfg.Settings.PullStrategy, "ff-only")
	}
	if got := len(cfg.Settings.Clean.ProtectedBranches); got < 4 {
		t.Errorf("expected at least 4 default protected branches, got %d", got)
	}
}

func TestSaveLoad_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := Default()
	cfg.Repos = []RepoEntry{{Path: "/tmp/a", Name: "a"}}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Repos) != 1 || got.Repos[0].Name != "a" {
		t.Errorf("roundtrip lost repo entry: %+v", got.Repos)
	}
}

func TestSave_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deeper", "config.yaml")
	if err := Save(path, Default()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not written: %v", err)
	}
}
