package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PublishOptions struct {
	Endpoint       string
	Organization   string
	RepositoryName string
	Namespace      string
	Name           string
	GPGKeyID       string
	Version        string
	WorkingDir     string
}

var publishOptions = &PublishOptions{}

var publishProviderCmd = &cobra.Command{
	Use:   "publish-provider",
	Short: "Publish provider to registry",
	Run: func(cmd *cobra.Command, args []string) {
		publishProvider(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(publishProviderCmd)

	publishProviderCmd.Flags().StringVar(&publishOptions.Endpoint, "endpoint", "", "Repository endpoint")
	publishProviderCmd.Flags().StringVar(&publishOptions.Organization, "organization", "", "Repository organization")
	publishProviderCmd.Flags().StringVar(&publishOptions.RepositoryName, "repository", "private", "Repository name")
	publishProviderCmd.Flags().StringVar(&publishOptions.Namespace, "namespace", "", "Provider namespace")
	publishProviderCmd.Flags().StringVar(&publishOptions.Name, "name", "", "Provider namespace")
	publishProviderCmd.Flags().StringVar(&publishOptions.GPGKeyID, "gpg-key-id", "", "GPG Key ID")
	publishProviderCmd.Flags().StringVar(&publishOptions.Version, "version", "", "Provider version")
	publishProviderCmd.Flags().StringVar(&publishOptions.WorkingDir, "working-dir", "", "Provider working directory")

	publishProviderCmd.MarkFlagRequired("endpoint")
	publishProviderCmd.MarkFlagRequired("organization")
	publishProviderCmd.MarkFlagRequired("name")
	publishProviderCmd.MarkFlagRequired("namespace")
	publishProviderCmd.MarkFlagRequired("gpg-key-id")
	publishProviderCmd.MarkFlagRequired("version")
}

func publishProvider(ctx context.Context) {
	providerRequest := models.RegistryProvidersRequest{
		Data: models.RegistryProvidersRequestData{
			Type: "registry-providers",
			Attributes: models.RegistryProvidersRequestAttributes{
				Name:         publishOptions.Name,
				Namespace:    publishOptions.Namespace,
				RegistryName: publishOptions.RepositoryName,
			},
		},
	}

	_, err := CreateProviderRequest(publishOptions.Endpoint, providerRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	providerVersionsRequest := models.RegistryProviderVersionsRequest{
		Data: models.RegistryProviderVersionsRequestData{
			Type: "registry-provider-versions",
			Attributes: models.RegistryProviderVersionsRequestAttributes{
				Version:   publishOptions.Version,
				KeyID:     publishOptions.GPGKeyID,
				Protocols: []string{"6.0"},
			},
		},
	}

	providerVersionsResponse, err := CreateProviderVersionRequest(publishOptions.Endpoint, providerVersionsRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	shaSumsFile := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", publishOptions.Name, publishOptions.Version)
	shaSumsSigFile := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS.sig", publishOptions.Name, publishOptions.Version)

	shaSumsPath := filepath.Join(publishOptions.WorkingDir, shaSumsFile)
	shaSumsSigPath := filepath.Join(publishOptions.WorkingDir, shaSumsSigFile)

	err = uploadFile(shaSumsPath, providerVersionsResponse.Data.Links.ShasumsUpload)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Uploaded SHA256SUMS file: ", shaSumsFile)
	err = uploadFile(shaSumsSigPath, providerVersionsResponse.Data.Links.ShasumsSigUpload)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Uploaded SHA256SUMS SIG file: ", shaSumsSigFile)

	content, err := readFileContents(shaSumsPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := parseSha256SUMS([]byte(content))
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range data {
		providerBinaryPath := filepath.Join(publishOptions.WorkingDir, k)

		_, err = os.Stat(providerBinaryPath)
		if !os.IsNotExist(err) {
			operatingSystem, architecture := parseProviderFile(k)

			platformRequest := models.RegistryProviderVersionPlatformsRequest{
				Data: models.RegistryProviderVersionPlatformsRequestData{
					Type: "registry-provider-version-platforms",
					Attributes: models.RegistryProviderVersionPlatformsRequestAttributes{
						OS:       operatingSystem,
						Arch:     architecture,
						Shasum:   v,
						Filename: k,
					},
				},
			}

			platformResponse, err := CreateProviderVersionPlatformsRequest(publishOptions.Endpoint, platformRequest)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = uploadFile(providerBinaryPath, platformResponse.Data.Links.ProviderBinaryUpload)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Uploaded provider binary file: ", k)
		} else {
			fmt.Println("Skipping provider binary file: ", k)
		}
	}
}

func CreateProviderRequest(endpoint string, request models.RegistryProvidersRequest) (*models.RegistryProvidersResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		os.Exit(1)
	}

	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers", publishOptions.Organization)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body:", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var response models.RegistryProvidersResponse
		err := json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}

	return nil, fmt.Errorf("Request failed with status %d:\n%s\n", resp.StatusCode, string(body))
}

func CreateProviderVersionRequest(endpoint string, request models.RegistryProviderVersionsRequest) (*models.RegistryProviderVersionsResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		os.Exit(1)
	}

	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s/versions", publishOptions.Organization, publishOptions.RepositoryName, publishOptions.Namespace, publishOptions.Name)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body:", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var response models.RegistryProviderVersionsResponse
		err := json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}

	return nil, fmt.Errorf("Request failed with status %d:\n%s\n", resp.StatusCode, string(body))
}

func CreateProviderVersionPlatformsRequest(endpoint string, request models.RegistryProviderVersionPlatformsRequest) (*models.RegistryProviderVersionPlatformsResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		os.Exit(1)
	}

	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s/versions/%s/platforms", publishOptions.Organization, publishOptions.RepositoryName, publishOptions.Namespace, publishOptions.Name, publishOptions.Version)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body:", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var response models.RegistryProviderVersionPlatformsResponse
		err := json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}

	return nil, fmt.Errorf("Request failed with status %d:\n%s\n", resp.StatusCode, string(body))
}

func uploadFile(filePath, url string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	fileContents, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(fileContents))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response: %w", err)
		}
		return fmt.Errorf(string(respBody))
	}

	return nil
}

func parseSha256SUMS(content []byte) (map[string]string, error) {
	dataMap := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) != 2 {
			continue
		}

		key := fields[1]
		value := fields[0]
		dataMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dataMap, nil
}

func parseProviderFile(name string) (string, string) {
	trimmed := strings.TrimSuffix(name, ".zip")
	parts := strings.Split(trimmed, "_")
	if len(parts) >= 3 {
		os := parts[len(parts)-2]
		arch := parts[len(parts)-1]
		return os, arch
	}

	return "", ""
}
