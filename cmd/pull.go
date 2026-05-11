package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/git"
	"github.com/stefanosbou/repokit/internal/registry"
	"github.com/stefanosbou/repokit/internal/runner"
)

var pullStrategy string
var pullFilter string
var pullForce bool

var errPullConflict = errors.New("merge conflict")

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull updates across registered repos in parallel",
	RunE: func(cmd *cobra.Command, args []string) error {
		strategy := pullStrategy
		if strategy == "" {
			strategy = globals.Cfg.Settings.PullStrategy
		}
		if strategy == "" {
			strategy = "ff-only"
		}
		switch strategy {
		case "ff-only", "rebase", "merge":
		default:
			return fmt.Errorf("invalid --strategy %q", strategy)
		}

		reg := registry.New(globals.Cfg)
		repos := reg.All()
		if globals.Repo != "" {
			r, ok := reg.ByName(globals.Repo)
			if !ok {
				return fmt.Errorf("unknown repo: %s", globals.Repo)
			}
			repos = []config.RepoEntry{r}
		}

		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		if pullFilter == "behind" {
			repos = filterRepos(ctx, repos, "behind", globals.Cfg.Settings.Clean.StaleAfterDays, globals.Parallel)
		}

		fmt.Printf("Pulling %d repos...\n\n", len(repos))

		tasks := make([]runner.Task, 0, len(repos))
		for _, r := range repos {
			tasks = append(tasks, runner.Task{
				RepoName: r.Name,
				RepoPath: r.Path,
				Run: func(ctx context.Context) runner.Result {
					st, err := git.Status(ctx, r.Path)
					if err != nil {
						return runner.Result{Status: runner.StatusError, Err: err}
					}
					if st.DirtyFiles > 0 && !pullForce {
						return runner.Result{Status: runner.StatusSkipped, Message: "uncommitted changes"}
					}
					res, err := git.Pull(ctx, r.Path, strategy)
					if err != nil {
						return runner.Result{Status: runner.StatusError, Err: err}
					}
					switch {
					case res.Conflict:
						return runner.Result{Status: runner.StatusError, Message: "Merge conflict", Err: errPullConflict}
					case res.UpToDate:
						return runner.Result{Status: runner.StatusOK, Message: "Already up to date"}
					case res.Updated:
						return runner.Result{Status: runner.StatusOK, Message: "Updated"}
					}
					return runner.Result{Status: runner.StatusOK, Message: res.Message}
				},
			})
		}

		results := runner.RunAll(ctx, tasks, globals.Parallel)

		var updated, skipped, conflicts, errs int
		for r := range results {
			printPullLine(r)

			switch {
			case errors.Is(r.Err, errPullConflict):
				conflicts++
			case r.Err != nil:
				errs++
			case r.Status == runner.StatusSkipped:
				skipped++
			default:
				updated++
			}
		}

		fmt.Printf("\n%d updated  ·  %d skipped  ·  %d conflict\n", updated, skipped, conflicts)
		if conflicts+errs > 0 {
			return fmt.Errorf("pull: %d conflicts, %d errors", conflicts, errs)
		}
		return nil
	},
}

func printPullLine(r runner.Result) {
	switch {
	case errors.Is(r.Err, errPullConflict):
		fmt.Printf("  %s  %-20s %s\n", color.RedString("✗"), r.RepoName, "Merge conflict")
	case r.Err != nil:
		fmt.Printf("  %s  %-20s %s\n", color.RedString("✗"), r.RepoName, r.Err.Error())
	case r.Status == runner.StatusSkipped:
		fmt.Printf("  %s  %-20s Skipped — %s\n", color.YellowString("⚠"), r.RepoName, r.Message)
	case r.Message == "Already up to date":
		fmt.Printf("  %s  %-20s Already up to date\n", color.GreenString("✓"), r.RepoName)
	default:
		fmt.Printf("  %s  %-20s %s\n", color.GreenString("↓"), r.RepoName, r.Message)
	}
}

// filterRepos returns repos whose derived state matches `want`, fetching
// statuses in parallel via runner.RunAll.
func filterRepos(ctx context.Context, repos []config.RepoEntry, want string, staleAfter int, parallel int) []config.RepoEntry {
	tasks := make([]runner.Task, 0, len(repos))
	for _, r := range repos {
		tasks = append(tasks, runner.Task{
			RepoName: r.Name,
			RepoPath: r.Path,
			Run: func(ctx context.Context) runner.Result {
				st, err := git.Status(ctx, r.Path)
				if err != nil {
					return runner.Result{Status: runner.StatusError, Err: err}
				}
				return runner.Result{Status: runner.StatusOK, Data: st}
			},
		})
	}
	byName := make(map[string]config.RepoEntry, len(repos))
	for _, r := range repos {
		byName[r.Name] = r
	}
	out := make([]config.RepoEntry, 0, len(repos))
	for r := range runner.RunAll(ctx, tasks, parallel) {
		if r.Err != nil {
			continue
		}
		if deriveState(r, staleAfter) == want {
			out = append(out, byName[r.RepoName])
		}
	}
	return out
}

func resultsToJSON(rs []runner.Result) []map[string]any {
	out := make([]map[string]any, 0, len(rs))
	for _, r := range rs {
		row := map[string]any{
			"repo":    r.RepoName,
			"path":    r.RepoPath,
			"status":  r.Status,
			"message": r.Message,
		}
		if r.Err != nil {
			row["error"] = r.Err.Error()
		}
		out = append(out, row)
	}
	return out
}

func init() {
	pullCmd.Flags().StringVar(&pullStrategy, "strategy", "", "Pull strategy: ff-only|rebase|merge (defaults to config)")
	pullCmd.Flags().StringVar(&pullFilter, "filter", "", "Only pull repos in this state (currently supports: behind)")
	pullCmd.Flags().BoolVar(&pullForce, "force", false, "Pull even repos with uncommitted changes")
	rootCmd.AddCommand(pullCmd)
}
