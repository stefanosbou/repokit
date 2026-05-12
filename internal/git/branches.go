package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MergedBranches returns branches whose tip is reachable from base, excluding base itself.
func MergedBranches(ctx context.Context, repoPath, base string) ([]Branch, error) {
	out, err := Run(ctx, repoPath, "for-each-ref", "--merged="+base,
		"--format=%(refname:short)|%(committerdate:unix)|%(HEAD)", "refs/heads")
	if err != nil {
		return nil, err
	}
	all := parseBranches(out)
	merged := all[:0]
	for _, b := range all {
		if b.Name != base {
			merged = append(merged, b)
		}
	}
	return merged, nil
}

// DeleteBranch deletes a local branch. force=true uses `-D`, otherwise `-d`.
func DeleteBranch(ctx context.Context, repoPath, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	if _, err := Run(ctx, repoPath, "branch", flag, branch); err != nil {
		return fmt.Errorf("delete branch %s: %w", branch, err)
	}
	return nil
}

func parseBranches(out string) []Branch {
	var bs []Branch
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 2 {
			continue
		}
		secs, _ := strconv.ParseInt(parts[1], 10, 64)
		b := Branch{
			Name:           parts[0],
			LastCommitTime: time.Unix(secs, 0),
		}
		if len(parts) == 3 && parts[2] == "*" {
			b.IsCurrent = true
		}
		bs = append(bs, b)
	}
	return bs
}
