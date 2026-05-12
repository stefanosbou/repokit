package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// makeRepoDir simulates a git repo by creating an empty .git directory.
func makeRepoDir(t *testing.T, root, rel string) string {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Join(full, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return full
}

func TestFindRepos_FindsAtMultipleDepths(t *testing.T) {
	root := t.TempDir()
	makeRepoDir(t, root, "a")
	makeRepoDir(t, root, "nested/b")
	makeRepoDir(t, root, "nested/deeper/c")

	got, err := FindRepos(root, 5)
	if err != nil {
		t.Fatalf("FindRepos: %v", err)
	}
	sort.Strings(got)
	want := []string{
		filepath.Join(root, "a"),
		filepath.Join(root, "nested/b"),
		filepath.Join(root, "nested/deeper/c"),
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestFindRepos_RespectsDepth(t *testing.T) {
	root := t.TempDir()
	makeRepoDir(t, root, "level1/level2/level3/repo")
	got, _ := FindRepos(root, 2)
	if len(got) != 0 {
		t.Errorf("depth=2 should not reach repo, got %v", got)
	}
	got, _ = FindRepos(root, 5)
	if len(got) != 1 {
		t.Errorf("depth=5 should find repo, got %v", got)
	}
}

func TestFindRepos_SkipsHiddenAndNodeModules(t *testing.T) {
	root := t.TempDir()
	makeRepoDir(t, root, ".cache/repo")
	makeRepoDir(t, root, "node_modules/dep")
	makeRepoDir(t, root, "real")
	got, _ := FindRepos(root, 5)
	if len(got) != 1 || filepath.Base(got[0]) != "real" {
		t.Errorf("expected only 'real', got %v", got)
	}
}

func TestFindRepos_DoesNotDescendIntoGitDir(t *testing.T) {
	root := t.TempDir()
	makeRepoDir(t, root, "outer")
	// Simulate nested .git/worktrees containing what looks like another repo.
	if err := os.MkdirAll(filepath.Join(root, "outer", ".git", "worktrees", "x", ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	got, _ := FindRepos(root, 5)
	if len(got) != 1 {
		t.Errorf("expected 1 repo, got %v", got)
	}
}
