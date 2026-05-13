package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stefanosbou/repokit/internal/runner"
)

// --- initialModel ---

func TestInitialModel_setsRepoSliceAndIndex(t *testing.T) {
	tasks := []runner.Task{
		{RepoName: "alpha", RepoPath: "/a"},
		{RepoName: "beta-long", RepoPath: "/b"},
	}
	ch := make(chan runner.Result, 2)
	m := initialModel(tasks, ch)

	if len(m.repos) != 2 {
		t.Fatalf("want 2 repos, got %d", len(m.repos))
	}
	if m.byName["alpha"] != 0 {
		t.Errorf("alpha: want index 0, got %d", m.byName["alpha"])
	}
	if m.byName["beta-long"] != 1 {
		t.Errorf("beta-long: want index 1, got %d", m.byName["beta-long"])
	}
	if m.nameWidth != 9 { // len("beta-long") == 9
		t.Errorf("nameWidth: want 9, got %d", m.nameWidth)
	}
	if m.total != 2 {
		t.Errorf("total: want 2, got %d", m.total)
	}
	if m.repos[0].done {
		t.Error("repos should start as not done")
	}
}

// --- Update: resultMsg ---

func TestUpdate_resultMsg_marksRepoDone(t *testing.T) {
	tasks := []runner.Task{
		{RepoName: "alpha", RepoPath: "/a"},
		{RepoName: "beta", RepoPath: "/b"},
	}
	ch := make(chan runner.Result, 2)
	m := initialModel(tasks, ch)

	newM, _ := m.Update(resultMsg{r: runner.Result{
		RepoName: "alpha",
		Status:   runner.StatusOK,
		Message:  "Updated",
	}})
	nm := newM.(model)

	if !nm.repos[0].done {
		t.Error("alpha should be marked done")
	}
	if nm.repos[0].status != runner.StatusOK {
		t.Errorf("want StatusOK, got %v", nm.repos[0].status)
	}
	if nm.repos[0].message != "Updated" {
		t.Errorf("want message 'Updated', got %q", nm.repos[0].message)
	}
	if nm.done != 1 {
		t.Errorf("done count: want 1, got %d", nm.done)
	}
	if nm.repos[1].done {
		t.Error("beta should still be pending")
	}
}

func TestUpdate_resultMsg_preservesError(t *testing.T) {
	tasks := []runner.Task{{RepoName: "repo", RepoPath: "/r"}}
	ch := make(chan runner.Result, 1)
	m := initialModel(tasks, ch)

	sentinel := context.DeadlineExceeded
	newM, _ := m.Update(resultMsg{r: runner.Result{
		RepoName: "repo",
		Status:   runner.StatusError,
		Err:      sentinel,
	}})
	nm := newM.(model)

	if nm.repos[0].err != sentinel {
		t.Errorf("want sentinel error, got %v", nm.repos[0].err)
	}
}

// --- Update: all done → tea.Quit ---

func TestUpdate_lastResult_returnsQuit(t *testing.T) {
	tasks := []runner.Task{{RepoName: "only", RepoPath: "/o"}}
	ch := make(chan runner.Result, 1)
	m := initialModel(tasks, ch)

	_, cmd := m.Update(resultMsg{r: runner.Result{
		RepoName: "only",
		Status:   runner.StatusOK,
	}})

	if cmd == nil {
		t.Fatal("expected a command (tea.Quit), got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// --- Update: doneMsg → tea.Quit ---

func TestUpdate_doneMsg_returnsQuit(t *testing.T) {
	tasks := []runner.Task{{RepoName: "repo", RepoPath: "/r"}}
	ch := make(chan runner.Result, 1)
	m := initialModel(tasks, ch)

	_, cmd := m.Update(doneMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// --- View ---

func TestView_pendingShowsDots(t *testing.T) {
	tasks := []runner.Task{{RepoName: "myrepo", RepoPath: "/r"}}
	ch := make(chan runner.Result, 1)
	m := initialModel(tasks, ch)

	view := m.View()
	if !strings.Contains(view, "myrepo") {
		t.Error("view should contain repo name")
	}
	if !strings.Contains(view, "···") {
		t.Error("pending repo should show ···")
	}
}

func TestView_doneShowsMessage(t *testing.T) {
	tasks := []runner.Task{{RepoName: "myrepo", RepoPath: "/r"}}
	ch := make(chan runner.Result, 1)
	m := initialModel(tasks, ch)
	// manually mark done
	m.repos[0].done = true
	m.repos[0].status = runner.StatusOK
	m.repos[0].message = "Updated"

	view := m.View()
	if !strings.Contains(view, "Updated") {
		t.Errorf("done repo view should contain message, got: %q", view)
	}
}

// --- RunWithProgress non-TTY path ---

func TestRunWithProgress_nonTTY_callsFallbackPerResult(t *testing.T) {
	// In test environments stdout is not a TTY, so non-TTY path runs automatically.
	tasks := []runner.Task{
		{
			RepoName: "repo-a",
			RepoPath: "/a",
			Run: func(ctx context.Context) runner.Result {
				return runner.Result{Status: runner.StatusOK, Message: "Fetched"}
			},
		},
		{
			RepoName: "repo-b",
			RepoPath: "/b",
			Run: func(ctx context.Context) runner.Result {
				return runner.Result{Status: runner.StatusError, Message: "boom"}
			},
		},
	}

	seen := map[string]bool{}
	fallback := func(r runner.Result) {
		seen[r.RepoName] = true
	}

	results, err := RunWithProgress(context.Background(), "Test title", tasks, 2, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
	if !seen["repo-a"] || !seen["repo-b"] {
		t.Errorf("fallback not called for all repos: seen=%v", seen)
	}
}

func TestRunWithProgress_nonTTY_returnsAllResults(t *testing.T) {
	tasks := []runner.Task{
		{
			RepoName: "ok-repo",
			RepoPath: "/ok",
			Run: func(ctx context.Context) runner.Result {
				return runner.Result{Status: runner.StatusOK, Message: "done"}
			},
		},
	}

	results, err := RunWithProgress(context.Background(), "T", tasks, 1, func(runner.Result) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].RepoName != "ok-repo" {
		t.Errorf("want RepoName ok-repo, got %q", results[0].RepoName)
	}
}
