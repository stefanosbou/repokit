package registry

import (
	"testing"

	"github.com/stefanosbou/repokit/internal/config"
)

func sampleConfig() *config.Config {
	c := config.Default()
	c.Repos = []config.RepoEntry{
		{Path: "/dev/api", Name: "api"},
		{Path: "/dev/web", Name: "web"},
		{Path: "/dev/utils", Name: "utils"},
	}
	return c
}

func TestAdd_NewEntry(t *testing.T) {
	c := config.Default()
	r := New(c)
	if !r.Add(config.RepoEntry{Path: "/dev/x", Name: "x"}) {
		t.Errorf("expected added=true on new entry")
	}
	if len(c.Repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(c.Repos))
	}
}

func TestAdd_DuplicatePath_Skipped(t *testing.T) {
	c := sampleConfig()
	r := New(c)
	if r.Add(config.RepoEntry{Path: "/dev/api", Name: "renamed"}) {
		t.Errorf("expected added=false for duplicate path")
	}
	if len(c.Repos) != 3 {
		t.Errorf("repo count changed: %d", len(c.Repos))
	}
}

func TestRemove_ByName(t *testing.T) {
	c := sampleConfig()
	r := New(c)
	if !r.Remove("web") {
		t.Errorf("expected Remove to return true")
	}
	if len(c.Repos) != 2 {
		t.Errorf("expected 2 repos after remove, got %d", len(c.Repos))
	}
}

func TestResolveByName(t *testing.T) {
	c := sampleConfig()
	r := New(c)
	got, ok := r.ByName("api")
	if !ok || got.Path != "/dev/api" {
		t.Errorf("ByName(api) = %+v, ok=%v", got, ok)
	}
	if _, ok := r.ByName("missing"); ok {
		t.Errorf("ByName(missing) should be not ok")
	}
}
