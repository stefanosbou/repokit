package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunAll_RunsAllTasks(t *testing.T) {
	var counter int32
	var tasks []Task
	for range 10 {
		tasks = append(tasks, Task{
			RepoName: "r",
			RepoPath: "/tmp/r",
			Run: func(ctx context.Context) Result {
				atomic.AddInt32(&counter, 1)
				return Result{Status: StatusOK}
			},
		})
	}
	results := RunAll(context.Background(), tasks, 4)
	count := 0
	for r := range results {
		if r.Status != StatusOK {
			t.Errorf("unexpected status: %+v", r)
		}
		count++
	}
	if count != 10 || atomic.LoadInt32(&counter) != 10 {
		t.Errorf("count=%d counter=%d, want 10/10", count, counter)
	}
}

func TestRunAll_BoundedConcurrency(t *testing.T) {
	var inflight, peak int32
	const max = 3
	var tasks []Task
	for range 12 {
		tasks = append(tasks, Task{
			Run: func(ctx context.Context) Result {
				cur := atomic.AddInt32(&inflight, 1)
				for {
					p := atomic.LoadInt32(&peak)
					if cur <= p || atomic.CompareAndSwapInt32(&peak, p, cur) {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt32(&inflight, -1)
				return Result{Status: StatusOK}
			},
		})
	}
	for range RunAll(context.Background(), tasks, max) {
	}
	if atomic.LoadInt32(&peak) > int32(max) {
		t.Errorf("peak concurrency %d exceeded max %d", peak, max)
	}
}

func TestRunAll_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var tasks []Task
	for range 8 {
		tasks = append(tasks, Task{
			Run: func(ctx context.Context) Result {
				select {
				case <-ctx.Done():
					return Result{Status: StatusError, Err: ctx.Err()}
				case <-time.After(time.Second):
					return Result{Status: StatusOK}
				}
			},
		})
	}
	results := RunAll(ctx, tasks, 2)
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	gotCancel := false
	for r := range results {
		if r.Err != nil && errors.Is(r.Err, context.Canceled) {
			gotCancel = true
		}
	}
	if !gotCancel {
		t.Error("expected at least one task to observe cancellation")
	}
}
