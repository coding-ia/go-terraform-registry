package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/client/api_client"
	"net/http"
)

type ModuleVersionDeleteOptions struct {
	Endpoint     string
	Organization string
	Registry     string
	Namespace    string
	Name         string
	Provider     string
	Version      string
}

var moduleVersionDeleteOptions = &ModuleVersionDeleteOptions{}

var moduleVersionDeleteCmd = &cobra.Command{
	Use:   "delete-module-version",
	Short: "Delete module version from registry",
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
		deleteModuleVersion(cmd.Context())
	},
}

func init() {
	moduleCmd.AddCommand(moduleVersionDeleteCmd)

	moduleVersionDeleteCmd.Flags().StringVar(&moduleVersionDeleteOptions.Endpoint, "endpoint", "", "Registry endpoint")
	moduleVersionDeleteCmd.Flags().StringVar(&moduleVersionDeleteOptions.Organization, "organization", "", "Registry organization")
	moduleVersionDeleteCmd.Flags().StringVar(&moduleVersionDeleteOptions.Registry, "registry", "private", "Registry name")
	moduleVersionDeleteCmd.Flags().StringVar(&moduleVersionDeleteOptions.Namespace, "namespace", "", "Module namespace")
	moduleVersionDeleteCmd.Flags().StringVar(&moduleVersionDeleteOptions.Name, "name", "", "Module namespace")
	moduleVersionDeleteCmd.Flags().StringVar(&moduleVersionDeleteOptions.Provider, "provider", "", "Module provider")
	moduleVersionDeleteCmd.Flags().StringVar(&moduleVersionDeleteOptions.Version, "version", "", "Module version")
	moduleVersionDeleteCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = publishModuleVersionCmd.MarkFlagRequired("endpoint")
	_ = publishModuleVersionCmd.MarkFlagRequired("organization")
	_ = publishModuleVersionCmd.MarkFlagRequired("name")
	_ = publishModuleVersionCmd.MarkFlagRequired("namespace")
	_ = publishModuleVersionCmd.MarkFlagRequired("provider")
	_ = publishModuleVersionCmd.MarkFlagRequired("version")
}

func deleteModuleVersion(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)

	statusCode, err := DeleteModuleVersionRequest(client, moduleVersionDeleteOptions.Endpoint)
	if err != nil {
		fmt.Println(fmt.Errorf("error getting provider request [%d]: %w", statusCode, err))
		return
	}

	if statusCode == http.StatusNoContent {
		fmt.Println(fmt.Sprintf("Module version %s deleted", moduleVersionDeleteOptions.Version))
	}
}

func DeleteModuleVersionRequest(client *api_client.APIClient, endpoint string) (int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-modules/%s/%s/%s/%s/%s", moduleVersionDeleteOptions.Organization, moduleVersionDeleteOptions.Registry, moduleVersionDeleteOptions.Namespace, moduleVersionDeleteOptions.Name, moduleVersionDeleteOptions.Provider, moduleVersionDeleteOptions.Version)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	statusCode, err := client.DeleteRequest(url)
	if err != nil {
		return statusCode, err
	}

	return statusCode, nil
}
