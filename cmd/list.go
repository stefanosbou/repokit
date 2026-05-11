package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/stefanosbou/repokit/internal/registry"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := registry.New(globals.Cfg)
		repos := reg.All()

		if len(repos) == 0 {
			fmt.Println("No repos registered.")
			return nil
		}
		// tw := output.NewTable(os.Stdout, []string{"NAME", "PATH", "GROUPS"})
		for _, r := range repos {
			// tw.Row(r.Name, r.Path, strings.Join(r.Groups, ", "))
			fmt.Printf("%s\t%s\n", r.Name, r.Path)
		}
		// _ = tw.Flush()
		fmt.Printf("\n%d repos registered.\n", len(repos))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
