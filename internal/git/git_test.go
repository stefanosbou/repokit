package git

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a fresh git repo in t.TempDir() with default branch "main".
// Returns the absolute repo path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
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

func TestRun_OK(t *testing.T) {
	repo := initRepo(t)
	out, err := Run(context.Background(), repo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(out) != "main" {
		t.Errorf("got branch %q, want main", out)
	}
}

func TestRun_Error(t *testing.T) {
	repo := initRepo(t)
	_, err := Run(context.Background(), repo, "this-is-not-a-command")
	if err == nil {
		t.Fatal("expected error from invalid git subcommand")
	}
}
