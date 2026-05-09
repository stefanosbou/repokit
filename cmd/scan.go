package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/scanner"
)

var scanDepth int

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Discover and register git repos under a directory",
	Long:  "Walks a directory tree, finds every .git, and registers any new repos in the config.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		root, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		fmt.Printf("Scanning %s...\n\n", root)

		paths, err := scanner.FindRepos(root, scanDepth)
		if err != nil {
			return err
		}

		var added []config.RepoEntry
		for _, p := range paths {
			name := filepath.Base(p)
			entry := config.RepoEntry{Path: p, Name: name}

			added = append(added, entry)
		}

		if len(added) == 0 {
			fmt.Println("No new repos found.")
			return nil
		}
		fmt.Printf("Found %d new repos:\n", len(added))
		for _, e := range added {
			fmt.Printf("  + %-20s %s\n", e.Name, e.Path)
		}
		fmt.Println()

		return nil
	},
}

func init() {
	scanCmd.Flags().IntVar(&scanDepth, "depth", 5, "Max directory depth to walk")
	rootCmd.AddCommand(scanCmd)
}
