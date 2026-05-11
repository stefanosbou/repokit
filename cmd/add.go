package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/registry"
)

var addName string

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Register a single repo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not a git repo: %s", path)
		} else if err != nil {
			return err
		}
		name := addName
		if name == "" {
			name = filepath.Base(path)
		}
		entry := config.RepoEntry{Path: path, Name: name}

		reg := registry.New(globals.Cfg)
		if !reg.Add(entry) {
			return fmt.Errorf("path already registered: %s", path)
		}

		if err := config.Save(globals.ConfigPath, globals.Cfg); err != nil {
			return err
		}

		fmt.Printf("✓ Registered: %s  %s\n", name, path)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addName, "name", "", "Override the registered name (default: directory name)")
	rootCmd.AddCommand(addCmd)
}
