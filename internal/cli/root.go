package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"os"
)

type AuthenticationOptions struct {
	Token string
}

var authenticationOptions = &AuthenticationOptions{}

var rootCmd = &cobra.Command{
	Use:   "tfrepoctl",
	Short: "A CLI interface for go-terraform-registry",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getURLHost(uri string) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

func addAuthFlag(cmd *cobra.Command, endpoint string) {
	host, _ := getURLHost(endpoint)
	token := os.Getenv(fmt.Sprintf("TF_TOKEN_%s", host))
	cmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", token, "Authorization token")
	if token == "" {
		_ = cmd.MarkFlagRequired("auth-token")
	}
}
