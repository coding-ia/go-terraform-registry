package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/models"
	"io"
	"net/http"
	"os"
)

type GPGOptions struct {
	Endpoint         string
	Namespace        string
	GPGPublicKeyPath string
}

var gpgOptions = &GPGOptions{}

var gpgCmd = &cobra.Command{
	Use:   "gpg",
	Short: "Manage GPG keys",
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a GPG key",
	Run: func(cmd *cobra.Command, args []string) {
		gpgAdd(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(gpgCmd)
	gpgCmd.AddCommand(addCmd)

	addCmd.Flags().StringVar(&gpgOptions.Endpoint, "endpoint", "", "Repository endpoint")
	addCmd.Flags().StringVar(&gpgOptions.Namespace, "namespace", "", "Provider namespace")
	addCmd.Flags().StringVar(&gpgOptions.GPGPublicKeyPath, "gpg-key-file", "", "GPG key file path")

	addCmd.MarkFlagRequired("endpoint")
	addCmd.MarkFlagRequired("namespace")
	addCmd.MarkFlagRequired("gpg-key-file")
}

func gpgAdd(ctx context.Context) {
	gpgKey, err := readFileContents(gpgOptions.GPGPublicKeyPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	gpgKeyRequest := models.GPGKeyRequest{
		Data: models.GPGKeyRequestData{
			Type: "gpg-keys",
			Attributes: models.GPGKeyRequestAttributes{
				Namespace:  gpgOptions.Namespace,
				AsciiArmor: gpgKey,
			},
		},
	}

	gpgKeyResponse, err := CreateGPGRequest(gpgOptions.Endpoint, gpgKeyRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("GPG key successfully created")
	fmt.Println(fmt.Sprintf("KEY ID: %s", gpgKeyResponse.Data.Attributes.KeyID))
}

func CreateGPGRequest(endpoint string, request models.GPGKeyRequest) (*models.GPGKeyResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		os.Exit(1)
	}

	apiEndpoint := "/api/registry/private/v2/gpg-keys"
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

	if resp.StatusCode == http.StatusCreated {
		var response models.GPGKeyResponse
		err := json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}

	return nil, fmt.Errorf("Request failed with status %d:\n%s\n", resp.StatusCode, string(body))
}

func readFileContents(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
