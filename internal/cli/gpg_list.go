package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/client/api_client"
	"net/url"
)

type GPGListOptions struct {
	Endpoint  string
	Namespace string
}

var gpgListOptions = &GPGListOptions{}

var gpgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List GPG key's",
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
		gpgList(cmd.Context())
	},
}

func init() {
	gpgCmd.AddCommand(gpgListCmd)

	gpgListCmd.Flags().StringVar(&gpgListOptions.Endpoint, "endpoint", "", "Repository endpoint")
	gpgListCmd.Flags().StringVar(&gpgListOptions.Namespace, "namespace", "", "Provider namespace")
	gpgListCmd.Flags().StringVar(&authenticationOptions.Token, "auth-token", "", "Authorization token")

	_ = gpgListCmd.MarkFlagRequired("endpoint")
	_ = gpgListCmd.MarkFlagRequired("namespace")
}

func gpgList(_ context.Context) {
	client := api_client.NewAPIClient(authenticationOptions.Token)

	gpgKeyListResponse, _, err := ListGPGRequest(client, gpgListOptions.Endpoint, gpgListOptions.Namespace)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("GPG Keys:")
	for _, gpgKeyData := range gpgKeyListResponse.Data {
		fmt.Println(gpgKeyData.Attributes.KeyID)
	}
}

func ListGPGRequest(client *api_client.APIClient, endpoint string, namespace string) (*models.GPGKeysListResponse, int, error) {
	apiEndpoint := "/api/registry/private/v2/gpg-keys"
	fullUrl := fmt.Sprintf("%s%s", endpoint, apiEndpoint)

	u, err := url.Parse(fullUrl)
	if err != nil {
		return nil, -1, err
	}

	q := u.Query()
	q.Set("filter[namespace]", namespace)
	u.RawQuery = q.Encode()

	var response models.GPGKeysListResponse
	statusCode, err := client.GetRequest(u.String(), &response)
	if err != nil {
		return nil, statusCode, err
	}

	return &response, statusCode, nil
}
