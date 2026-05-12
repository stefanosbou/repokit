package git

import (
	"context"
	"testing"
)

func TestListBranches(t *testing.T) {
	repo := initRepo(t)
	gitInRepo(t, repo, "branch", "feat/x")
	gitInRepo(t, repo, "branch", "feat/y")
	bs, err := ListBranches(context.Background(), repo)
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	names := map[string]bool{}
	for _, b := range bs {
		names[b.Name] = true
	}
	for _, want := range []string{"main", "feat/x", "feat/y"} {
		if !names[want] {
			t.Errorf("missing branch %q in %+v", want, bs)
		}
	}
}

func TestMergedBranches(t *testing.T) {
	repo := initRepo(t)
	gitInRepo(t, repo, "checkout", "-q", "-b", "feat/done")
	gitInRepo(t, repo, "commit", "--allow-empty", "-q", "-m", "done")
	gitInRepo(t, repo, "checkout", "-q", "main")
	gitInRepo(t, repo, "merge", "-q", "feat/done")
	gitInRepo(t, repo, "checkout", "-q", "-b", "feat/wip")
	gitInRepo(t, repo, "commit", "--allow-empty", "-q", "-m", "wip")
	gitInRepo(t, repo, "checkout", "-q", "main")

	merged, err := MergedBranches(context.Background(), repo, "main")
	if err != nil {
		t.Fatalf("MergedBranches: %v", err)
	}
	gotNames := map[string]bool{}
	for _, b := range merged {
		gotNames[b.Name] = true
	}
	if !gotNames["feat/done"] {
		t.Errorf("expected feat/done in merged, got %v", merged)
	}
	if gotNames["feat/wip"] {
		t.Errorf("feat/wip should not be merged")
	}
	if gotNames["main"] {
		t.Errorf("base branch should be excluded from results")
	}
}

func TestDeleteBranch_Safe(t *testing.T) {
	repo := initRepo(t)
	gitInRepo(t, repo, "checkout", "-q", "-b", "feat/done")
	gitInRepo(t, repo, "commit", "--allow-empty", "-q", "-m", "done")
	gitInRepo(t, repo, "checkout", "-q", "main")
	gitInRepo(t, repo, "merge", "-q", "feat/done")

	if err := DeleteBranch(context.Background(), repo, "feat/done", false); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}
}

func TestDeleteBranch_RefusesUnmerged_WithoutForce(t *testing.T) {
	repo := initRepo(t)
	gitInRepo(t, repo, "checkout", "-q", "-b", "feat/wip")
	gitInRepo(t, repo, "commit", "--allow-empty", "-q", "-m", "wip")
	gitInRepo(t, repo, "checkout", "-q", "main")

	if err := DeleteBranch(context.Background(), repo, "feat/wip", false); err == nil {
		t.Fatal("expected error deleting unmerged branch without force")
	}
}
