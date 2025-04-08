package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	apimodels "go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/client/api_client"
	"net/http"
)

type ModuleCreateOptions struct {
	Endpoint     string
	Organization string
	Registry     string
	Namespace    string
	Name         string
	Provider     string
}

var moduleOptions = &ModuleCreateOptions{}

var createModuleCmd = &cobra.Command{
	Use:   "create-module",
	Short: "Create module in registry",
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
		createModule(cmd.Context())
	},
}

func init() {
	moduleCmd.AddCommand(createModuleCmd)

	createModuleCmd.Flags().StringVar(&moduleOptions.Endpoint, "endpoint", "", "Registry endpoint")
	createModuleCmd.Flags().StringVar(&moduleOptions.Organization, "organization", "", "Registry organization")
	createModuleCmd.Flags().StringVar(&moduleOptions.Registry, "registry", "private", "Registry name")
	createModuleCmd.Flags().StringVar(&moduleOptions.Namespace, "namespace", "", "Module namespace")
	createModuleCmd.Flags().StringVar(&moduleOptions.Name, "name", "", "Module namespace")
	createModuleCmd.Flags().StringVar(&moduleOptions.Provider, "provider", "", "Module provider")
	createModuleCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = createModuleCmd.MarkFlagRequired("endpoint")
	_ = createModuleCmd.MarkFlagRequired("organization")
	_ = createModuleCmd.MarkFlagRequired("name")
	_ = createModuleCmd.MarkFlagRequired("namespace")
	_ = createModuleCmd.MarkFlagRequired("provider")
}

func createModule(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)

	moduleRequest := apimodels.ModulesRequest{
		Data: apimodels.ModulesDataRequest{
			Type: "registry-modules",
			Attributes: apimodels.ModulesAttributesRequest{
				Name:         moduleOptions.Name,
				Namespace:    moduleOptions.Namespace,
				Provider:     moduleOptions.Provider,
				RegistryName: moduleOptions.Registry,
				NoCode:       true,
			},
		},
	}

	m, statusCode, err := CreateModuleRequest(client, moduleOptions.Endpoint, moduleRequest)
	if err != nil && statusCode != http.StatusNotFound {
		fmt.Println(fmt.Errorf("error getting provider request [%d]: %w", statusCode, err))
		return
	}
	if m == nil {
		fmt.Println(fmt.Sprintf("Error creating module %s\\%s: %v", moduleOptions.Namespace, moduleOptions.Name, err))
		return
	}

	if statusCode == http.StatusCreated {
		fmt.Println(fmt.Sprintf("Module %s\\%s\\%s created", m.Data.Attributes.Provider, m.Data.Attributes.Namespace, m.Data.Attributes.Name))
	}
}

func CreateModuleRequest(client *api_client.APIClient, endpoint string, request apimodels.ModulesRequest) (*apimodels.ModulesResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-modules", moduleOptions.Organization)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ModulesResponse
	statusCode, err := client.PostRequest(url, request, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}
