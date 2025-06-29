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

type ModuleVersionPublishOptions struct {
	Endpoint     string
	Organization string
	Registry     string
	Namespace    string
	Name         string
	Provider     string
	Version      string
	CommitSHA    string
	File         string
	ChunkUpload  bool
}

var moduleVersionPublishOptions = &ModuleVersionPublishOptions{}

var publishModuleVersionCmd = &cobra.Command{
	Use:   "publish-module-version",
	Short: "Publish module version to registry",
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
		publishModuleVersion(cmd.Context())
	},
}

func init() {
	moduleCmd.AddCommand(publishModuleVersionCmd)

	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.Endpoint, "endpoint", "", "Registry endpoint")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.Organization, "organization", "", "Registry organization")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.Registry, "registry", "private", "Registry name")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.Namespace, "namespace", "", "Module namespace")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.Name, "name", "", "Module namespace")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.Provider, "provider", "", "Module provider")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.Version, "version", "", "Module version")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.CommitSHA, "commit-sha", "", "Module commit SHA")
	publishModuleVersionCmd.Flags().StringVar(&moduleVersionPublishOptions.File, "archive-file", "", "Module archive file [tar.gz]")
	publishModuleVersionCmd.Flags().BoolVar(&moduleVersionPublishOptions.ChunkUpload, "chunk-upload", false, "Upload chunks")
	publishModuleVersionCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = publishModuleVersionCmd.MarkFlagRequired("endpoint")
	_ = publishModuleVersionCmd.MarkFlagRequired("organization")
	_ = publishModuleVersionCmd.MarkFlagRequired("name")
	_ = publishModuleVersionCmd.MarkFlagRequired("namespace")
	_ = publishModuleVersionCmd.MarkFlagRequired("provider")
	_ = publishModuleVersionCmd.MarkFlagRequired("version")
	_ = publishModuleVersionCmd.MarkFlagRequired("commit-sha")
}

func publishModuleVersion(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)

	m, statusCode, err := GetModulesRequest(client, moduleVersionPublishOptions.Endpoint)
	if err != nil && statusCode != http.StatusNotFound {
		fmt.Println(fmt.Errorf("error getting provider request [%d]: %w", statusCode, err))
		return
	}
	if m == nil && statusCode != http.StatusOK {
		fmt.Println(fmt.Sprintf("Module %s\\%s\\%s does not exist: %v", moduleVersionPublishOptions.Provider, moduleVersionPublishOptions.Namespace, moduleVersionPublishOptions.Name, err))
		return
	}

	moduleVersionRequest := apimodels.ModuleVersionsRequest{
		Data: apimodels.ModuleVersionsDataRequest{
			Type: "registry-module-versions",
			Attributes: apimodels.ModuleVersionsAttributesRequest{
				Version:   moduleVersionPublishOptions.Version,
				CommitSHA: moduleVersionPublishOptions.CommitSHA,
			},
		},
	}

	v, statusCode, err := CreateModuleVersionsRequest(client, moduleVersionPublishOptions.Endpoint, moduleVersionRequest)
	if err != nil && statusCode != http.StatusNotFound {
		fmt.Println(fmt.Errorf("error getting provider request [%d]: %w", statusCode, err))
		return
	}
	if v == nil {
		fmt.Println(fmt.Sprintf("Error creating module version %s: %v", moduleVersionPublishOptions.Version, err))
		return
	}

	if statusCode == http.StatusCreated {
		if !moduleVersionPublishOptions.ChunkUpload {
			err = uploadFile(moduleVersionPublishOptions.File, v.Data.Links.Upload)
		} else {
			err = uploadFileChunks(moduleVersionPublishOptions.File, v.Data.Links.Upload)
		}
		if err != nil {
			fmt.Println(fmt.Errorf("error uploading file [%s]: %w", moduleVersionPublishOptions.File, err))
			return
		}
		fmt.Println(fmt.Sprintf("Module version %s created", v.Data.Attributes.Version))
	}
}

func GetModulesRequest(client *api_client.APIClient, endpoint string) (*apimodels.ModulesResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-modules/%s/%s/%s/%s", moduleVersionPublishOptions.Organization, moduleVersionPublishOptions.Registry, moduleVersionPublishOptions.Namespace, moduleVersionPublishOptions.Name, moduleVersionPublishOptions.Provider)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ModulesResponse
	statusCode, err := client.GetRequest(url, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}

func CreateModuleVersionsRequest(client *api_client.APIClient, endpoint string, request apimodels.ModuleVersionsRequest) (*apimodels.ModuleVersionsResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-modules/%s/%s/%s/%s/versions", moduleVersionPublishOptions.Organization, moduleVersionPublishOptions.Registry, moduleVersionPublishOptions.Namespace, moduleVersionPublishOptions.Name, moduleVersionPublishOptions.Provider)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ModuleVersionsResponse
	statusCode, err := client.PostRequest(url, request, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}
