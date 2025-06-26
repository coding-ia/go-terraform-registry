package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/client/api_client"
)

type ProviderVersionListOptions struct {
	Endpoint     string
	Organization string
	Registry     string
	Namespace    string
	Name         string
}

var providerVersionListOptions = &ProviderVersionListOptions{}

var providerVersionsListCmd = &cobra.Command{
	Use:   "list-provider-versions",
	Short: "List all provider versions in the registry",
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
		listProviderVersion(cmd.Context())
	},
}

func init() {
	providerCmd.AddCommand(providerVersionsListCmd)

	providerVersionsListCmd.Flags().StringVar(&providerVersionListOptions.Endpoint, "endpoint", "", "Registry endpoint")
	providerVersionsListCmd.Flags().StringVar(&providerVersionListOptions.Organization, "organization", "", "Registry organization")
	providerVersionsListCmd.Flags().StringVar(&providerVersionListOptions.Registry, "registry", "private", "Registry name")
	providerVersionsListCmd.Flags().StringVar(&providerVersionListOptions.Namespace, "namespace", "", "Provider namespace")
	providerVersionsListCmd.Flags().StringVar(&providerVersionListOptions.Name, "name", "", "Provider namespace")
	providerVersionsListCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = providerVersionsListCmd.MarkFlagRequired("endpoint")
	_ = providerVersionsListCmd.MarkFlagRequired("organization")
	_ = providerVersionsListCmd.MarkFlagRequired("name")
	_ = providerVersionsListCmd.MarkFlagRequired("namespace")
}

func listProviderVersion(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)

	providerVersionsListResponse, _, err := ListProviderVersionsRequest(client, providerVersionListOptions.Endpoint)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Provider versions:")
	for _, providerVersionData := range providerVersionsListResponse.Data {
		fmt.Println(providerVersionData.Attributes.Version)
	}
}

func ListProviderVersionsRequest(client *api_client.APIClient, endpoint string) (*models.ProviderVersionsListResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s/versions", providerVersionListOptions.Organization, providerVersionListOptions.Registry, providerVersionListOptions.Namespace, providerVersionListOptions.Name)
	fullUrl := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response models.ProviderVersionsListResponse
	statusCode, err := client.GetRequest(fullUrl, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}
