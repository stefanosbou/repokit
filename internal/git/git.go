package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Branch is shared between branches.go and status.go.
type Branch struct {
	Name           string
	IsCurrent      bool
	LastCommitTime time.Time
}

// RepoStatus is the parsed result of `git status --porcelain` plus context.
type RepoStatus struct {
	Branch      string
	Detached    bool
	DirtyFiles  int
	Ahead       int
	Behind      int
	HasUpstream bool
	LastCommit  time.Time
}

// PullResult describes the outcome of a pull operation.
type PullResult struct {
	UpToDate bool
	Updated  bool
	Conflict bool
	Commits  int    // commits pulled in (when known)
	Message  string // human-readable summary
}

// Run executes `git <args...>` inside repoPath and returns combined stdout.
// Stderr is included in the returned error on failure.
func Run(ctx context.Context, repoPath string, args ...string) (string, error) {
	full := append([]string{"-C", repoPath}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("git %s (%s): %w", strings.Join(args, " "), strings.TrimSpace(stderr.String()), err)
	}
	return stdout.String(), nil
}
