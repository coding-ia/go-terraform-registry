package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/client/api_client"
	"net/http"
)

type ProviderVersionDeleteOptions struct {
	Endpoint     string
	Organization string
	Registry     string
	Namespace    string
	Name         string
	Provider     string
	Version      string
}

var providerVersionDeleteOptions = &ProviderVersionDeleteOptions{}

var providerVersionDeleteCmd = &cobra.Command{
	Use:   "delete-provider-version",
	Short: "Delete provider version from registry",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		authToken, _ := cmd.Flags().GetString("auth-token")
		if authToken == "" && !setAuthTokenFromEnv(endpoint) {
			_ = authToken
			return errors.New("required flag(s) \"auth-token\" not set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		deleteProviderVersion(cmd.Context())
	},
}

func init() {
	providerCmd.AddCommand(providerVersionDeleteCmd)

	providerVersionDeleteCmd.Flags().StringVar(&providerVersionDeleteOptions.Endpoint, "endpoint", "", "Registry endpoint")
	providerVersionDeleteCmd.Flags().StringVar(&providerVersionDeleteOptions.Organization, "organization", "", "Registry organization")
	providerVersionDeleteCmd.Flags().StringVar(&providerVersionDeleteOptions.Registry, "registry", "private", "Registry name")
	providerVersionDeleteCmd.Flags().StringVar(&providerVersionDeleteOptions.Namespace, "namespace", "", "Module namespace")
	providerVersionDeleteCmd.Flags().StringVar(&providerVersionDeleteOptions.Name, "name", "", "Module namespace")
	providerVersionDeleteCmd.Flags().StringVar(&providerVersionDeleteOptions.Provider, "provider", "", "Module provider")
	providerVersionDeleteCmd.Flags().StringVar(&providerVersionDeleteOptions.Version, "version", "", "Module version")
	providerVersionDeleteCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = providerVersionDeleteCmd.MarkFlagRequired("endpoint")
	_ = providerVersionDeleteCmd.MarkFlagRequired("organization")
	_ = providerVersionDeleteCmd.MarkFlagRequired("name")
	_ = providerVersionDeleteCmd.MarkFlagRequired("namespace")
	_ = providerVersionDeleteCmd.MarkFlagRequired("provider")
	_ = providerVersionDeleteCmd.MarkFlagRequired("version")
}

func deleteProviderVersion(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)

	statusCode, err := DeleteProviderVersionRequest(client, providerVersionDeleteOptions.Endpoint)
	if err != nil {
		fmt.Println(fmt.Errorf("error getting provider request [%d]: %w", statusCode, err))
		return
	}

	if statusCode == http.StatusNoContent {
		fmt.Println(fmt.Sprintf("Provider version %s deleted", providerVersionDeleteOptions.Version))
	}
}

func DeleteProviderVersionRequest(client *api_client.APIClient, endpoint string) (int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s/%s/%s", providerVersionDeleteOptions.Organization, providerVersionDeleteOptions.Registry, providerVersionDeleteOptions.Namespace, providerVersionDeleteOptions.Name, providerVersionDeleteOptions.Provider, providerVersionDeleteOptions.Version)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	statusCode, err := client.DeleteRequest(url)
	if err != nil {
		return statusCode, err
	}

	return statusCode, nil
}
