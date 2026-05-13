package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/git"
	"github.com/stefanosbou/repokit/internal/registry"
	"github.com/stefanosbou/repokit/internal/runner"
	"github.com/stefanosbou/repokit/internal/tui"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch updates across registered repos in parallel",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		tasks := make([]runner.Task, 0, len(repos))
		for _, r := range repos {
			tasks = append(tasks, runner.Task{
				RepoName: r.Name,
				RepoPath: r.Path,
				Run: func(ctx context.Context) runner.Result {
					if err := git.Fetch(ctx, r.Path); err != nil {
						return runner.Result{Status: runner.StatusError, Err: err}
					}
					return runner.Result{Status: runner.StatusOK, Message: "Fetched"}
				},
			})
		}

		printFetch := func(r runner.Result) {
			switch {
			case r.Err != nil:
				fmt.Printf("  %s  %-20s %s\n", color.RedString("✗"), r.RepoName, r.Err.Error())
			default:
				fmt.Printf("  %s  %-20s %s\n", color.GreenString("✓"), r.RepoName, r.Message)
			}
		}

		collected, err := tui.RunWithProgress(
			ctx,
			fmt.Sprintf("Fetching %d repos...", len(repos)),
			tasks,
			globals.Parallel,
			printFetch,
		)
		if err != nil {
			return err
		}

		var ok, errs int
		for _, r := range collected {
			if r.Err != nil {
				errs++
			} else {
				ok++
			}
		}

		fmt.Printf("\n%d fetched  ·  %d errors\n", ok, errs)
		if errs > 0 {
			return fmt.Errorf("fetch: %d errors", errs)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
