package cli

import (
	"github.com/spf13/cobra"
)

var gpgCmd = &cobra.Command{
	Use:   "gpg",
	Short: "Manage GPG keys",
}

func init() {
	rootCmd.AddCommand(gpgCmd)
}
