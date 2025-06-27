package main

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func fetchItems() []string {
	// In a real app, fetch from DB, API, etc.
	return []string{"apple", "banana", "cherry", "date"}
}

var rootCmd = &cobra.Command{
	Use:   "carapace-dynamic",
	Short: "Demo CLI with dynamic carapace completion",
}

var listCmd = &cobra.Command{
	Use:   "list [item]",
	Short: "List objects from the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("You selected: %s\n", args[0])
		return nil
	},
}

func init() {
	carapace.Gen(rootCmd).Standalone() // enables _carapace and disables legacy completion

	carapace.Gen(listCmd).PositionalCompletion(
		carapace.ActionCallback(func(ctx carapace.Context) carapace.Action {
			items := fetchItems()
			return carapace.ActionValues(items...)
		}),
	)

	rootCmd.AddCommand(listCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
