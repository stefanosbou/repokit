package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scanDepth int

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Discover and register git repos under a directory",
	Long:  "Walks a directory tree, finds every .git, and registers any new repos in the config.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("scan called")
	},
}

func init() {
	scanCmd.Flags().IntVar(&scanDepth, "depth", 5, "Max directory depth to walk")
	rootCmd.AddCommand(scanCmd)
}
