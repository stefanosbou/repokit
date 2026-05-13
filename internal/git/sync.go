package git

import (
	"context"
	"fmt"
	"strings"
)

// Fetch runs `git fetch --prune` in repoPath.
func Fetch(ctx context.Context, repoPath string) error {
	_, err := Run(ctx, repoPath, "fetch", "--prune")
	return err
}

// Pull runs `git pull` with the requested strategy: "ff-only", "rebase", or "merge".
func Pull(ctx context.Context, repoPath, strategy string) (*PullResult, error) {
	args := []string{"pull"}
	switch strategy {
	case "ff-only", "":
		args = append(args, "--ff-only")
	case "rebase":
		args = append(args, "--rebase")
	case "merge":
		args = append(args, "--no-rebase")
	default:
		return nil, fmt.Errorf("unknown pull strategy %q", strategy)
	}
	out, err := Run(ctx, repoPath, args...)
	res := &PullResult{Message: strings.TrimSpace(out)}
	if err != nil {
		if hasUnmergedEntries(ctx, repoPath) {
			res.Conflict = true
			return res, nil
		}
		return res, err
	}
	// LC_ALL=C guarantees "Already up to date." (no hyphen variant).
	if strings.Contains(out, "Already up to date") {
		res.UpToDate = true
		return res, nil
	}
	res.Updated = true
	return res, nil
}

// hasUnmergedEntries reports whether the repo has unmerged files after a failed pull.
func hasUnmergedEntries(ctx context.Context, repoPath string) bool {
	out, err := Run(ctx, repoPath, "status", "--porcelain=v2")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "u ") {
			return true
		}
	}
	return false
}
