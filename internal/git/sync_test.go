package git

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

// fixtureCloneWithUpstream returns (clone, upstream) where the clone is one commit behind.
func fixtureCloneWithUpstream(t *testing.T) (string, string) {
	t.Helper()
	upstream := initRepo(t)
	clone := t.TempDir()
	clone, _ = filepath.EvalSymlinks(clone)
	if out, err := exec.Command("git", "clone", "-q", upstream, clone).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}
	gitInRepo(t, clone, "config", "user.email", "test@example.com")
	gitInRepo(t, clone, "config", "user.name", "test")
	// Make the upstream advance by one commit so clone is behind.
	gitInRepo(t, upstream, "config", "receive.denyCurrentBranch", "ignore")
	gitInRepo(t, upstream, "commit", "--allow-empty", "-q", "-m", "remote-1")
	return clone, upstream
}

func TestFetch(t *testing.T) {
	clone, _ := fixtureCloneWithUpstream(t)
	if err := Fetch(context.Background(), clone); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	// After fetch, ahead/behind reflects new state.
	ahead, behind, err := AheadBehind(context.Background(), clone)
	if err != nil {
		t.Fatalf("AheadBehind: %v", err)
	}
	if ahead != 0 || behind != 1 {
		t.Errorf("after fetch ahead=%d behind=%d, want 0/1", ahead, behind)
	}
}

func TestPull_FastForward(t *testing.T) {
	clone, _ := fixtureCloneWithUpstream(t)
	res, err := Pull(context.Background(), clone, "ff-only")
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if !res.Updated {
		t.Errorf("expected Updated=true, got %+v", res)
	}
}

func TestPull_AlreadyUpToDate(t *testing.T) {
	upstream := initRepo(t)
	clone := t.TempDir()
	clone, _ = filepath.EvalSymlinks(clone)
	if out, err := exec.Command("git", "clone", "-q", upstream, clone).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}
	gitInRepo(t, clone, "config", "user.email", "test@example.com")
	gitInRepo(t, clone, "config", "user.name", "test")
	res, err := Pull(context.Background(), clone, "ff-only")
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if !res.UpToDate {
		t.Errorf("expected UpToDate=true, got %+v", res)
	}
}
