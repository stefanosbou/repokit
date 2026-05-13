package cmd

import (
	"context"
	"fmt"

	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/registry"
	"github.com/stefanosbou/repokit/internal/runner"
)

func selectRepos(reg *registry.Registry, only string) ([]config.RepoEntry, error) {
	if only == "" {
		return reg.All(), nil
	}
	r, ok := reg.ByName(only)
	if !ok {
		return nil, fmt.Errorf("unknown repo: %s", only)
	}
	return []config.RepoEntry{r}, nil
}

func buildTasks(repos []config.RepoEntry, op func(context.Context, config.RepoEntry) runner.Result) []runner.Task {
	tasks := make([]runner.Task, 0, len(repos))
	for _, r := range repos {
		tasks = append(tasks, runner.Task{
			RepoName: r.Name,
			RepoPath: r.Path,
			Run:      func(ctx context.Context) runner.Result { return op(ctx, r) },
		})
	}
	return tasks
}
