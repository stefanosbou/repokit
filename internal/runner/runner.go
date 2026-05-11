package runner

import (
	"context"
	"runtime"
	"sync"
)

// Status is the lifecycle outcome a task reports back through Result.
type Status string

const (
	StatusOK      Status = "ok"
	StatusError   Status = "error"
	StatusSkipped Status = "skipped"
)

type Task struct {
	RepoName string
	RepoPath string
	Run      func(ctx context.Context) Result
}

type Result struct {
	RepoName string
	RepoPath string
	Status   Status
	Message  string
	Data     any
	Err      error
}

// RunAll executes tasks in parallel up to maxWorkers (0 → runtime.NumCPU()).
// Results are streamed to the returned channel as they complete; the channel
// is closed when all tasks finish or the context is cancelled.
func RunAll(ctx context.Context, tasks []Task, maxWorkers int) <-chan Result {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxWorkers > len(tasks) && len(tasks) > 0 {
		maxWorkers = len(tasks)
	}
	out := make(chan Result, len(tasks))
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				out <- Result{RepoName: t.RepoName, RepoPath: t.RepoPath, Status: StatusError, Err: ctx.Err()}
				return
			}
			defer func() { <-sem }()
			res := t.Run(ctx)
			res.RepoName = t.RepoName
			res.RepoPath = t.RepoPath
			if res.Err != nil && res.Status == "" {
				res.Status = StatusError
			}
			out <- res
		}(t)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
