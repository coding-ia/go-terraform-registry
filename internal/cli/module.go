package cli

import "github.com/spf13/cobra"

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Manage modules",
}

func init() {
	rootCmd.AddCommand(moduleCmd)
}
