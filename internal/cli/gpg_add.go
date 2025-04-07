package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/client/api_client"
	"os"
)

type GPGOptions struct {
	Endpoint         string
	Namespace        string
	GPGPublicKeyPath string
}

var gpgOptions = &GPGOptions{}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a GPG key",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		value := setAuthTokenFlag(cmd, endpoint)

		if value == "" {
			return errors.New("required flag(s) \"auth-token\" not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		gpgAdd(cmd.Context())
	},
}

func init() {
	gpgCmd.AddCommand(addCmd)

	addCmd.Flags().StringVar(&gpgOptions.Endpoint, "endpoint", "", "Repository endpoint")
	addCmd.Flags().StringVar(&gpgOptions.Namespace, "namespace", "", "Provider namespace")
	addCmd.Flags().StringVar(&gpgOptions.GPGPublicKeyPath, "gpg-key-file", "", "GPG key file path")
	addCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = addCmd.MarkFlagRequired("endpoint")
	_ = addCmd.MarkFlagRequired("namespace")
	_ = addCmd.MarkFlagRequired("gpg-key-file")
}

func gpgAdd(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)
	gpgKey, err := readFileContents(gpgOptions.GPGPublicKeyPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	gpgKeyRequest := models.GPGKeysRequest{
		Data: models.GPGKeysDataRequest{
			Type: "gpg-keys",
			Attributes: models.GPGKeysAttributesRequest{
				Namespace:  gpgOptions.Namespace,
				AsciiArmor: gpgKey,
			},
		},
	}

	gpgKeyResponse, _, err := CreateGPGRequest(client, gpgOptions.Endpoint, gpgKeyRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("GPG key successfully created")
	fmt.Println(fmt.Sprintf("KEY ID: %s", gpgKeyResponse.Data.Attributes.KeyID))
}

func CreateGPGRequest(client *api_client.APIClient, endpoint string, request models.GPGKeysRequest) (*models.GPGKeysResponse, int, error) {
	apiEndpoint := "/api/registry/private/v2/gpg-keys"
	url := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	var response models.GPGKeysResponse
	statusCode, err := client.PostRequest(url, request, &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}

func readFileContents(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
