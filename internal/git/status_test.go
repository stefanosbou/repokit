package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func gitInRepo(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestCurrentBranch(t *testing.T) {
	repo := initRepo(t)
	got, err := CurrentBranch(context.Background(), repo)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if got != "main" {
		t.Errorf("got %q, want main", got)
	}
}

func TestStatus_Dirty(t *testing.T) {
	repo := initRepo(t)
	writeFile(t, repo, "a.txt", "hello")
	writeFile(t, repo, "b.txt", "world")
	st, err := Status(context.Background(), repo)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.DirtyFiles != 2 {
		t.Errorf("dirty files = %d, want 2", st.DirtyFiles)
	}
	if st.Branch != "main" {
		t.Errorf("branch = %q, want main", st.Branch)
	}
}

func TestStatus_Clean(t *testing.T) {
	repo := initRepo(t)
	st, err := Status(context.Background(), repo)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.DirtyFiles != 0 {
		t.Errorf("dirty files = %d, want 0", st.DirtyFiles)
	}
}

func TestLastCommitTime(t *testing.T) {
	repo := initRepo(t)
	got, err := LastCommitTime(context.Background(), repo)
	if err != nil {
		t.Fatalf("LastCommitTime: %v", err)
	}
	if time.Since(got) > time.Minute {
		t.Errorf("commit time too old: %v", got)
	}
}

func TestAheadBehind_Linked(t *testing.T) {
	upstream := initRepo(t)
	clone := t.TempDir()
	clone, _ = filepath.EvalSymlinks(clone)
	if out, err := exec.Command("git", "clone", "-q", upstream, clone).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}
	gitInRepo(t, clone, "config", "user.email", "test@example.com")
	gitInRepo(t, clone, "config", "user.name", "test")

	// 2 local commits ahead.
	gitInRepo(t, clone, "commit", "--allow-empty", "-q", "-m", "local-1")
	gitInRepo(t, clone, "commit", "--allow-empty", "-q", "-m", "local-2")
	// 1 remote commit behind.
	gitInRepo(t, upstream, "config", "receive.denyCurrentBranch", "ignore")
	gitInRepo(t, upstream, "commit", "--allow-empty", "-q", "-m", "remote-1")
	gitInRepo(t, clone, "fetch", "-q")

	ahead, behind, err := AheadBehind(context.Background(), clone)
	if err != nil {
		t.Fatalf("AheadBehind: %v", err)
	}
	if ahead != 2 || behind != 1 {
		t.Errorf("ahead=%d behind=%d, want 2/1", ahead, behind)
	}
}

func TestAheadBehind_NoUpstream(t *testing.T) {
	repo := initRepo(t)
	_, _, err := AheadBehind(context.Background(), repo)
	if err != ErrNoUpstream {
		t.Errorf("err = %v, want ErrNoUpstream", err)
	}
}

// TestAheadBehind_BrokenUpstream covers the case where branch.<name>.remote
// and branch.<name>.merge are configured, but the remote does not exist and
// no remote-tracking ref has been created. git emits an unusual stderr that
// substring matching missed; the explicit upstream probe must catch it.
func TestAheadBehind_BrokenUpstream(t *testing.T) {
	repo := initRepo(t)
	gitInRepo(t, repo, "config", "branch.main.remote", "origin")
	gitInRepo(t, repo, "config", "branch.main.merge", "refs/heads/main")

	_, _, err := AheadBehind(context.Background(), repo)
	if !errors.Is(err, ErrNoUpstream) {
		t.Errorf("err = %v, want ErrNoUpstream", err)
	}
}
