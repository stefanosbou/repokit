package cmd

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stefanosbou/repokit/internal/config"
)

// makeGitRepo creates a temp dir with an initialized git repo.
func makeGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"commit", "--allow-empty", "-q", "-m", "init"},
	}
	for _, args := range cmds {
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

// makeConfig writes a config with the given repos to a temp file and returns its path.
func makeConfig(t *testing.T, repos []config.RepoEntry) string {
	t.Helper()
	cfg := config.Default()
	cfg.Repos = repos
	p := filepath.Join(t.TempDir(), "config.yaml")
	if err := config.Save(p, cfg); err != nil {
		t.Fatalf("makeConfig: %v", err)
	}
	return p
}

// runCmd resets shared state and executes the root command with given args.
func runCmd(args ...string) error {
	*globals = Globals{}
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(context.Background())
}

func TestRemove_UnknownRepo(t *testing.T) {
	cfgPath := makeConfig(t, nil)
	err := runCmd("--config", cfgPath, "remove", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown repo")
	}
	if !strings.Contains(err.Error(), "unknown repo") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRemove_KnownRepo(t *testing.T) {
	repo := makeGitRepo(t)
	cfgPath := makeConfig(t, []config.RepoEntry{{Name: "myrepo", Path: repo}})
	if err := runCmd("--config", cfgPath, "remove", "myrepo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Repos) != 0 {
		t.Errorf("expected 0 repos after remove, got %d", len(cfg.Repos))
	}
}

func TestAdd_NewRepo(t *testing.T) {
	repo := makeGitRepo(t)
	cfgPath := makeConfig(t, nil)
	if err := runCmd("--config", cfgPath, "add", repo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Repos) != 1 {
		t.Errorf("expected 1 repo after add, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0].Name != filepath.Base(repo) {
		t.Errorf("repo name = %q, want %q", cfg.Repos[0].Name, filepath.Base(repo))
	}
}

func TestAdd_DuplicatePath(t *testing.T) {
	repo := makeGitRepo(t)
	cfgPath := makeConfig(t, []config.RepoEntry{{Name: "myrepo", Path: repo}})
	err := runCmd("--config", cfgPath, "add", repo)
	if err == nil {
		t.Fatal("expected error for duplicate path")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	cfgPath := makeConfig(t, nil)
	if err := runCmd("--config", cfgPath, "list"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_WithRepos(t *testing.T) {
	repo := makeGitRepo(t)
	cfgPath := makeConfig(t, []config.RepoEntry{{Name: "myrepo", Path: repo}})
	if err := runCmd("--config", cfgPath, "list"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatus_WithRepo(t *testing.T) {
	repo := makeGitRepo(t)
	cfgPath := makeConfig(t, []config.RepoEntry{{Name: "myrepo", Path: repo}})
	if err := runCmd("--config", cfgPath, "status"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatus_UnknownRepo(t *testing.T) {
	cfgPath := makeConfig(t, nil)
	err := runCmd("--config", cfgPath, "--repo", "nonexistent", "status")
	if err == nil {
		t.Fatal("expected error for unknown repo filter")
	}
	if !strings.Contains(err.Error(), "unknown repo") {
		t.Errorf("unexpected error message: %v", err)
	}
}
