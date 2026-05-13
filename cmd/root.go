package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stefanosbou/repokit/internal/config"
)

// Globals is populated from persistent flags before any RunE runs.
type Globals struct {
	Repo       string
	Parallel   int
	ConfigPath string

	Cfg *config.Config
}

var globals = &Globals{}

var rootCmd = &cobra.Command{
	Use:           "repokit",
	Short:         "Manage a fleet of local git repositories",
	Long:          "repokit is a CLI for bulk-managing local git repos: registry, status dashboard, parallel pull/fetch, dead-branch reaper, and secret scanner.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(globals.ConfigPath)
		if err != nil {
			return err
		}
		globals.Cfg = cfg
		if !cmd.Root().PersistentFlags().Changed("parallel") && cfg.Settings.Parallel > 0 {
			globals.Parallel = cfg.Settings.Parallel
		}
		// First-run init: persist the default config so users have something to edit.
		if globals.ConfigPath == "" {
			path, _ := config.ConfigPath()
			if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
				_ = config.Save(path, cfg)
			}
		}
		return nil
	},
}

// Execute is the entry point invoked by main.
func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&globals.Repo, "repo", "", "Target a single repo by name")
	pf.IntVar(&globals.Parallel, "parallel", 0, "Max concurrent operations (0 = NumCPU)")
	pf.StringVar(&globals.ConfigPath, "config", "", "Override config file path")
}
