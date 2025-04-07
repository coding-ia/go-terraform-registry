package cli

import "github.com/spf13/cobra"

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manager providers",
}

func init() {
	rootCmd.AddCommand(providerCmd)
}
