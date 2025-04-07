package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"strings"
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
	host := parsedURL.Host
	host = strings.ReplaceAll(host, ":", "_")
	host = strings.ReplaceAll(host, ".", "_")
	host = strings.ReplaceAll(host, "-", "_")
	return host, nil
}

func setAuthTokenFlag(cmd *cobra.Command, endpoint string) string {
	value, _ := cmd.Flags().GetString("auth-token")
	if value == "" {
		host, _ := getURLHost(endpoint)
		tfToken := os.Getenv(fmt.Sprintf("TF_TOKEN_%s", host))
		_ = cmd.Flags().Set("auth-token", tfToken)
		return tfToken
	}
	return value
}
