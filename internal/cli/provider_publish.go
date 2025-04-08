package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	apimodels "go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/client/api_client"
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
	ChunkUpload    bool
}

var publishOptions = &PublishOptions{}

var publishProviderCmd = &cobra.Command{
	Use:   "publish-provider",
	Short: "Publish provider to registry",
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
		publishProvider(cmd.Context())
	},
}

const chunkSize = 1024 * 1024 // 1MB

func init() {
	providerCmd.AddCommand(publishProviderCmd)

	publishProviderCmd.Flags().StringVar(&publishOptions.Endpoint, "endpoint", "", "Repository endpoint")
	publishProviderCmd.Flags().StringVar(&publishOptions.Organization, "organization", "", "Repository organization")
	publishProviderCmd.Flags().StringVar(&publishOptions.RepositoryName, "repository", "private", "Repository name")
	publishProviderCmd.Flags().StringVar(&publishOptions.Namespace, "namespace", "", "Provider namespace")
	publishProviderCmd.Flags().StringVar(&publishOptions.Name, "name", "", "Provider namespace")
	publishProviderCmd.Flags().StringVar(&publishOptions.GPGKeyID, "gpg-key-id", "", "GPG Key ID")
	publishProviderCmd.Flags().StringVar(&publishOptions.Version, "version", "", "Provider version")
	publishProviderCmd.Flags().StringVar(&publishOptions.WorkingDir, "working-dir", "", "Provider working directory")
	publishProviderCmd.Flags().BoolVar(&publishOptions.ChunkUpload, "chunk-upload", false, "Upload chunks")
	publishProviderCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = publishProviderCmd.MarkFlagRequired("endpoint")
	_ = publishProviderCmd.MarkFlagRequired("organization")
	_ = publishProviderCmd.MarkFlagRequired("name")
	_ = publishProviderCmd.MarkFlagRequired("namespace")
	_ = publishProviderCmd.MarkFlagRequired("gpg-key-id")
	_ = publishProviderCmd.MarkFlagRequired("version")
}

func publishProvider(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)

	providerRequest := apimodels.ProvidersRequest{
		Data: apimodels.ProvidersDataRequest{
			Type: "registry-providers",
			Attributes: apimodels.ProvidersAttributesRequest{
				Name:         publishOptions.Name,
				Namespace:    publishOptions.Namespace,
				RegistryName: publishOptions.RepositoryName,
			},
		},
	}

	p, statusCode, err := GetProviderRequest(client, publishOptions.Endpoint)
	if err != nil && statusCode != http.StatusNotFound {
		fmt.Println(fmt.Errorf("error getting provider request [%d]: %w", statusCode, err))
		return
	}

	if p == nil {
		fmt.Println(fmt.Sprintf("Creating provider %s\\%s", publishOptions.Namespace, publishOptions.Name))
		_, _, err = CreateProviderRequest(client, publishOptions.Endpoint, providerRequest)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	providerVersionsRequest := apimodels.ProviderVersionsRequest{
		Data: apimodels.ProviderVersionsDataRequest{
			Type: "registry-provider-versions",
			Attributes: apimodels.ProviderVersionsAttributesRequest{
				Version:   publishOptions.Version,
				KeyID:     publishOptions.GPGKeyID,
				Protocols: []string{"6.0"},
			},
		},
	}

	pv, statusCode, err := GetProviderVersionRequest(client, publishOptions.Endpoint)
	if err != nil && statusCode != http.StatusNotFound {
		fmt.Println(fmt.Errorf("error getting provider request [%d]: %w", statusCode, err))
		return
	}
	if pv != nil {
		fmt.Println(fmt.Sprintf("Provider version %s [%s\\%s] already exists", publishOptions.Version, publishOptions.Namespace, publishOptions.Name))
		return
	}

	providerVersionsResponse, statusCode, err := CreateProviderVersionRequest(client, publishOptions.Endpoint, providerVersionsRequest)
	if err != nil {
		fmt.Println(fmt.Errorf("error getting provider version request [%d]: %w", statusCode, err))
		return
	}

	shaSumsFile := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", publishOptions.Name, publishOptions.Version)
	shaSumsSigFile := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS.sig", publishOptions.Name, publishOptions.Version)

	shaSumsPath := filepath.Join(publishOptions.WorkingDir, shaSumsFile)
	shaSumsSigPath := filepath.Join(publishOptions.WorkingDir, shaSumsSigFile)

	err = uploadFile(shaSumsPath, *providerVersionsResponse.Data.Links.ShasumsUpload)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Uploaded SHA256SUMS file: ", shaSumsFile)
	err = uploadFile(shaSumsSigPath, *providerVersionsResponse.Data.Links.ShasumsSigUpload)
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

			platformRequest := apimodels.ProviderVersionPlatformsRequest{
				Data: apimodels.ProviderVersionPlatformsDataRequest{
					Type: "registry-provider-version-platforms",
					Attributes: apimodels.ProviderVersionPlatformsAttributesRequest{
						OS:       operatingSystem,
						Arch:     architecture,
						Shasum:   v,
						Filename: k,
					},
				},
			}

			platformResponse, _, err := CreateProviderVersionPlatformsRequest(client, publishOptions.Endpoint, platformRequest)
			if err != nil {
				fmt.Println(err)
				return
			}

			if !publishOptions.ChunkUpload {
				err = uploadFile(providerBinaryPath, platformResponse.Data.Links.ProviderBinaryUpload)
			} else {
				err = uploadFileChunks(providerBinaryPath, platformResponse.Data.Links.ProviderBinaryUpload)
			}
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

func GetProviderRequest(client *api_client.APIClient, endpoint string) (*apimodels.ProvidersResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s", publishOptions.Organization, publishOptions.RepositoryName, publishOptions.Namespace, publishOptions.Name)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ProvidersResponse
	statusCode, err := client.GetRequest(url, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}

func CreateProviderRequest(client *api_client.APIClient, endpoint string, request apimodels.ProvidersRequest) (*apimodels.ProvidersResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers", publishOptions.Organization)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ProvidersResponse
	statusCode, err := client.PostRequest(url, request, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}

func GetProviderVersionRequest(client *api_client.APIClient, endpoint string) (*apimodels.ProviderVersionsResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s/versions/%s", publishOptions.Organization, publishOptions.RepositoryName, publishOptions.Namespace, publishOptions.Name, publishOptions.Version)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ProviderVersionsResponse
	statusCode, err := client.GetRequest(url, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}

func CreateProviderVersionRequest(client *api_client.APIClient, endpoint string, request apimodels.ProviderVersionsRequest) (*apimodels.ProviderVersionsResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s/versions", publishOptions.Organization, publishOptions.RepositoryName, publishOptions.Namespace, publishOptions.Name)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ProviderVersionsResponse
	statusCode, err := client.PostRequest(url, request, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}

func CreateProviderVersionPlatformsRequest(client *api_client.APIClient, endpoint string, request apimodels.ProviderVersionPlatformsRequest) (*apimodels.ProviderVersionPlatformsResponse, int, error) {
	apiEndpoint := fmt.Sprintf("/api/v2/organizations/%s/registry-providers/%s/%s/%s/versions/%s/platforms", publishOptions.Organization, publishOptions.RepositoryName, publishOptions.Namespace, publishOptions.Name, publishOptions.Version)
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response apimodels.ProviderVersionPlatformsResponse
	statusCode, err := client.PostRequest(url, request, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
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

func uploadFileChunks(filePath, url string) error {
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

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	totalChunks := (fileSize + chunkSize - 1) / chunkSize

	buf := make([]byte, chunkSize)
	for i := 0; i < int(totalChunks); i++ {
		bytesRead, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		chunk := buf[:bytesRead]

		err = sendChunk(url, chunk, i+1, int(totalChunks), fileInfo.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func sendChunk(uploadURL string, chunk []byte, chunkNumber int, totalChunks int, fileName string) error {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(chunk))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Chunk-Number", fmt.Sprintf("%d", chunkNumber))
	req.Header.Set("Total-Chunks", fmt.Sprintf("%d", totalChunks))
	req.Header.Set("File-Name", fileName)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
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
		operatingSystem := parts[len(parts)-2]
		architecture := parts[len(parts)-1]
		return operatingSystem, architecture
	}

	return "", ""
}
