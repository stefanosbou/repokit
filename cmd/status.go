package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/git"
	"github.com/stefanosbou/repokit/internal/output"
	"github.com/stefanosbou/repokit/internal/registry"
	"github.com/stefanosbou/repokit/internal/runner"
)

var statusFilter string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show fleet health dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := registry.New(globals.Cfg)
		repos, err := selectRepos(reg, globals.Repo)
		if err != nil {
			return err
		}
		ctx := cmd.Context()

		tasks := buildTasks(repos, func(ctx context.Context, r config.RepoEntry) runner.Result {
			st, err := git.Status(ctx, r.Path)
			if err != nil {
				return runner.Result{Status: runner.StatusError, Err: err}
			}
			return runner.Result{Status: runner.StatusOK, Data: st}
		})

		results := runner.RunAll(ctx, tasks, globals.Parallel)
		var collected []runner.Result
		for r := range results {
			collected = append(collected, r)
		}
		return renderStatus(collected, globals.Cfg.Settings.Clean.StaleAfterDays)
	},
}

func renderStatus(results []runner.Result, staleAfterDays int) error {
	return renderStatusTable(results, staleAfterDays)
}

func deriveState(r runner.Result, staleAfterDays int) string {
	if r.Err != nil {
		return "error"
	}
	st, ok := r.Data.(*git.RepoStatus)
	if !ok || st == nil {
		return "error"
	}
	if st.Detached {
		return "detached"
	}
	if st.DirtyFiles > 0 {
		return "dirty"
	}
	if !st.HasUpstream {
		return "no-upstream"
	}
	if st.Behind > 0 {
		return "behind"
	}
	if st.Ahead > 0 {
		return "unpushed"
	}
	if staleAfterDays > 0 && time.Since(st.LastCommit) > time.Duration(staleAfterDays)*24*time.Hour {
		return "stale"
	}
	return "clean"
}

func renderStatusTable(results []runner.Result, staleAfterDays int) error {
	fmt.Printf("Fleet: %d repos\n\n", len(results))
	tw := output.NewTable(os.Stdout, []string{"NAME", "BRANCH", "STATUS", "LAST COMMIT"})
	for _, r := range results {
		state := deriveState(r, staleAfterDays)
		if statusFilter != "" && state != statusFilter {
			continue
		}
		st, _ := r.Data.(*git.RepoStatus)
		branch := "-"
		statusCell := state
		var last string
		if st != nil {
			branch = st.Branch
			statusCell = formatStatusCell(state, st)
			last = output.RelTime(time.Since(st.LastCommit))
		}
		if r.Err != nil {
			statusCell = color.RedString("✗ error: %s", truncate(r.Err.Error(), 40))
		}
		tw.Row(r.RepoName, branch, statusCell, last)
	}
	_ = tw.Flush()
	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────")
	return summariseStatus(results, staleAfterDays)
}

func formatStatusCell(state string, st *git.RepoStatus) string {
	switch state {
	case "clean":
		return color.GreenString("✓ clean")
	case "dirty":
		return color.YellowString("⚠ dirty (+%d)", st.DirtyFiles)
	case "unpushed":
		return color.YellowString("⚠ unpushed (↑%d)", st.Ahead)
	case "behind":
		return color.RedString("✗ behind (↓%d)", st.Behind)
	case "detached":
		return color.CyanString("~ detached")
	case "stale":
		return color.HiBlackString("⊘ stale")
	case "no-upstream":
		return "? no upstream"
	}
	return state
}

func summariseStatus(results []runner.Result, staleAfterDays int) error {
	counts := map[string]int{}
	for _, r := range results {
		counts[deriveState(r, staleAfterDays)]++
	}
	fmt.Printf("%d clean  ·  %d warnings  ·  %d behind  ·  %d stale\n",
		counts["clean"], counts["dirty"]+counts["unpushed"], counts["behind"], counts["stale"])
	if counts["error"] > 0 {
		fmt.Fprintf(os.Stderr, "\n%d repos errored.\n", counts["error"])
		return errors.New("status: one or more repos failed")
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func init() {
	statusCmd.Flags().StringVar(&statusFilter, "filter", "", "Show only repos in this state (dirty|behind|stale|clean|unpushed|detached)")
	rootCmd.AddCommand(statusCmd)
}
