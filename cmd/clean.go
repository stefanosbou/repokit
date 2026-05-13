package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/git"
	"github.com/stefanosbou/repokit/internal/output"
	"github.com/stefanosbou/repokit/internal/registry"
	"github.com/stefanosbou/repokit/internal/runner"
)

var (
	cleanBase      string
	cleanOlderThan int
	cleanForce     bool
)

// hardProtected branches are never deletable, even if removed from config.
var hardProtected = []string{"main", "master", "develop", "staging"}

type repoPlan struct {
	repo     config.RepoEntry
	branches []git.Branch
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Cleanup operations across registered repos",
}

var cleanBranchesCmd = &cobra.Command{
	Use:   "branches",
	Short: "Delete merged local branches across all repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := registry.New(globals.Cfg)
		repos, err := selectRepos(reg, globals.Repo)
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		fmt.Printf("Scanning %d repos for merged branches...\n\n", len(repos))

		plans := make([]repoPlan, 0, len(repos))
		tasks := make([]runner.Task, 0, len(repos))
		for _, r := range repos {
			tasks = append(tasks, runner.Task{
				RepoName: r.Name,
				RepoPath: r.Path,
				Run: func(ctx context.Context) runner.Result {
					base := cleanBase
					if base == "" {
						b, err := detectBase(ctx, r.Path)
						if err != nil {
							return runner.Result{Status: runner.StatusError, Err: err}
						}
						base = b
					}
					all, err := git.MergedBranches(ctx, r.Path, base)
					if err != nil {
						return runner.Result{Status: runner.StatusError, Err: err}
					}
					eligible := filterDeletable(all, globals.Cfg.Settings.Clean.ProtectedBranches, cleanOlderThan)
					return runner.Result{Status: runner.StatusOK, Data: eligible, Message: base}
				},
			})
		}
		for r := range runner.RunAll(ctx, tasks, globals.Parallel) {
			if r.Err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s: %s\n", r.RepoName, r.Err)
				continue
			}
			bs, _ := r.Data.([]git.Branch)
			if len(bs) == 0 {
				continue
			}
			repo, _ := reg.ByName(r.RepoName)
			plans = append(plans, repoPlan{repo: repo, branches: bs})
		}

		total := 0
		for _, p := range plans {
			total += len(p.branches)
		}

		// Human preview.
		for _, p := range plans {
			fmt.Printf("  %s\n", p.repo.Name)
			for _, b := range p.branches {
				ago := output.RelTime(time.Since(b.LastCommitTime))
				fmt.Printf("    - %-30s merged %s\n", b.Name, ago)
			}
			fmt.Println()
		}
		fmt.Println("──────────────────────────────────────────────────")
		fmt.Printf("%d branches across %d repos eligible for deletion.\n", total, len(plans))

		if total == 0 {
			return nil
		}

		fmt.Printf("\nDelete all? [y/N]: ")
		r := bufio.NewReader(os.Stdin)
		line, _ := r.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "y") {
			fmt.Println("Aborted.")
			return nil
		}

		var deleteTasks []runner.Task
		for _, p := range plans {
			for _, b := range p.branches {
				deleteTasks = append(deleteTasks, runner.Task{
					RepoName: p.repo.Name,
					RepoPath: p.repo.Path,
					Run: func(ctx context.Context) runner.Result {
						if err := git.DeleteBranch(ctx, p.repo.Path, b.Name, cleanForce); err != nil {
							return runner.Result{Status: runner.StatusError, Err: err, Message: b.Name}
						}
						return runner.Result{Status: runner.StatusOK, Message: b.Name}
					},
				})
			}
		}
		var deleted, errs int
		for r := range runner.RunAll(ctx, deleteTasks, globals.Parallel) {
			if r.Err != nil {
				fmt.Printf("  %s  %-15s %-20s %s\n", color.RedString("✗"), r.RepoName, r.Message, r.Err.Error())
				errs++
			} else {
				fmt.Printf("  %s  %-15s %-20s deleted\n", color.GreenString("✓"), r.RepoName, r.Message)
				deleted++
			}
		}

		fmt.Printf("\n%d branches deleted.\n", deleted)
		if errs > 0 {
			return fmt.Errorf("clean branches: %d errors", errs)
		}
		return nil
	},
}

func detectBase(ctx context.Context, repoPath string) (string, error) {
	for _, candidate := range []string{"main", "master"} {
		if _, err := git.Run(ctx, repoPath, "rev-parse", "--verify", candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no main or master branch found")
}

func filterDeletable(bs []git.Branch, configProtected []string, olderThanDays int) []git.Branch {
	protected := map[string]bool{}
	for _, p := range hardProtected {
		protected[p] = true
	}
	for _, p := range configProtected {
		protected[p] = true
	}
	out := make([]git.Branch, 0, len(bs))
	for _, b := range bs {
		if b.IsCurrent || protected[b.Name] {
			continue
		}
		if olderThanDays > 0 && time.Since(b.LastCommitTime) < time.Duration(olderThanDays)*24*time.Hour {
			continue
		}
		out = append(out, b)
	}
	return out
}

func init() {
	cleanBranchesCmd.Flags().StringVar(&cleanBase, "base", "", "Base branch to check merged-into (default: auto-detect main/master)")
	cleanBranchesCmd.Flags().IntVar(&cleanOlderThan, "older-than", 0, "Only delete branches older than N days")
	cleanBranchesCmd.Flags().BoolVar(&cleanForce, "force", false, "Use git branch -D instead of -d")
	cleanCmd.AddCommand(cleanBranchesCmd)
	rootCmd.AddCommand(cleanCmd)
}
