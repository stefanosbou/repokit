package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/stefanosbou/repokit/internal/config"
	"github.com/stefanosbou/repokit/internal/registry"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Unregister a repo by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		reg := registry.New(globals.Cfg)
		if !reg.Remove(name) {
			return fmt.Errorf("unknown repo: %s", name)
		}
		if err := config.Save(globals.ConfigPath, globals.Cfg); err != nil {
			return err
		}
		fmt.Printf("✓ Removed: %s\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
