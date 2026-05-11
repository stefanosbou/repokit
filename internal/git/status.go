package git

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

// ErrNoUpstream is returned by AheadBehind when no tracking branch is set.
var ErrNoUpstream = errors.New("no upstream configured")

// CurrentBranch returns the short ref name of HEAD, or "HEAD" if detached.
func CurrentBranch(ctx context.Context, repoPath string) (string, error) {
	out, err := Run(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Status fetches the current branch + porcelain status + last commit time.
// AheadBehind is queried best-effort; if there is no upstream, HasUpstream is false.
func Status(ctx context.Context, repoPath string) (*RepoStatus, error) {
	branch, err := CurrentBranch(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	out, err := Run(ctx, repoPath, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	dirty := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			dirty++
		}
	}
	last, err := LastCommitTime(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	st := &RepoStatus{
		Branch:     branch,
		Detached:   branch == "HEAD",
		DirtyFiles: dirty,
		LastCommit: last,
	}
	ahead, behind, err := AheadBehind(ctx, repoPath)
	if err == nil {
		st.HasUpstream = true
		st.Ahead = ahead
		st.Behind = behind
	} else if !errors.Is(err, ErrNoUpstream) {
		return nil, err
	}
	return st, nil
}

// AheadBehind returns commits ahead and behind the upstream of HEAD.
// Returns ErrNoUpstream when no upstream is configured or its tracking ref
// cannot be resolved (e.g. the remote was removed or never fetched).
func AheadBehind(ctx context.Context, repoPath string) (int, int, error) {
	// Probe upstream existence explicitly. `--quiet` makes git exit non-zero
	// without printing to stderr when @{u} is unset or unresolvable, which
	// avoids brittle stderr substring matching.
	probe, err := Run(ctx, repoPath, "rev-parse", "--symbolic-full-name", "--quiet", "@{u}")
	if err != nil || strings.TrimSpace(probe) == "" {
		return 0, 0, ErrNoUpstream
	}
	out, err := Run(ctx, repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return 0, 0, errors.New("unexpected rev-list output: " + out)
	}
	ahead, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	behind, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return ahead, behind, nil
}

// LastCommitTime returns the committer time of HEAD.
func LastCommitTime(ctx context.Context, repoPath string) (time.Time, error) {
	out, err := Run(ctx, repoPath, "log", "-1", "--format=%ct")
	if err != nil {
		return time.Time{}, err
	}
	secs, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(secs, 0), nil
}
