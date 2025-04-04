package dynamodb_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"strings"
)

var _ backend.RegistryBackend = &DynamoDBBackend{}

func (d *DynamoDBBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters, userParameters registrytypes.UserParameters) (*models.TerraformProviderPlatformResponse, error) {
	key := fmt.Sprintf("%s:%s:%s/%s", userParameters.Organization, "private", parameters.Namespace, parameters.Name)
	provider, err := getProvider(ctx, d.client, d.Tables.ProviderTableName, key)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, nil
	}

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.Tables.ProviderVersionTableName),
		KeyConditionExpression: aws.String("provider_id = :id and #v = :v"),
		ExpressionAttributeNames: map[string]string{
			"#v": "version",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: provider.ID},
			":v":  &types.AttributeValueMemberS{Value: parameters.Version},
		},
	}

	resp, err := d.client.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 1 {
		protocols := resp.Items[0]["protocols"].(*types.AttributeValueMemberSS).Value
		gpgKeyId := resp.Items[0]["gpg_key_id"].(*types.AttributeValueMemberS).Value
		gpgAsciiArmor := resp.Items[0]["gpg_ascii_armor"].(*types.AttributeValueMemberS).Value

		response := &models.TerraformProviderPlatformResponse{
			Protocols: protocols,
			SigningKeys: models.SigningKeys{
				GPGPublicKeys: []models.GPGPublicKeys{
					{
						KeyId:      gpgKeyId,
						AsciiArmor: gpgAsciiArmor,
					},
				},
			},
		}

		platforms := resp.Items[0]["platforms"].(*types.AttributeValueMemberL)
		for _, platform := range platforms.Value {
			p := platform.(*types.AttributeValueMemberS)

			var platform ProviderPlatform
			err := json.Unmarshal([]byte(p.Value), &platform)
			if err != nil {
				return nil, err
			}
			if strings.EqualFold(platform.OS, parameters.OS) &&
				strings.EqualFold(platform.Arch, parameters.Architecture) {
				response.Filename = platform.Filename
				response.Shasum = platform.SHASum
				response.OS = platform.OS
				response.Arch = platform.Arch

				return response, nil
			}
		}
	}

	return nil, nil
}

func (d *DynamoDBBackend) GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters, userParameters registrytypes.UserParameters) (*models.TerraformAvailableProvider, error) {
	key := fmt.Sprintf("%s:%s:%s/%s", userParameters.Organization, "private", parameters.Namespace, parameters.Name)
	provider, err := getProvider(ctx, d.client, d.Tables.ProviderTableName, key)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, nil
	}

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.Tables.ProviderVersionTableName),
		KeyConditionExpression: aws.String("provider_id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: provider.ID},
		},
	}

	resp, err := d.client.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 0 {
		return nil, nil
	}

	var versions []models.TerraformAvailableVersion
	for _, item := range resp.Items {
		version := item["version"].(*types.AttributeValueMemberS).Value
		protocols := item["protocols"].(*types.AttributeValueMemberSS).Value

		v := models.TerraformAvailableVersion{
			Version:   version,
			Protocols: protocols,
		}

		platforms := item["platforms"].(*types.AttributeValueMemberL)
		for _, platform := range platforms.Value {
			p := platform.(*types.AttributeValueMemberS)

			var platform ProviderPlatform
			err := json.Unmarshal([]byte(p.Value), &platform)
			if err != nil {
				return nil, err
			}
			v.Platforms = append(v.Platforms, models.TerraformAvailablePlatform{
				OS:   platform.OS,
				Arch: platform.Arch,
			})
		}
		versions = append(versions, v)
	}

	response := &models.TerraformAvailableProvider{
		Versions: versions,
	}

	return response, nil
}

func (d *DynamoDBBackend) GetModuleVersions(ctx context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error) {
	return nil, nil
}

func (d *DynamoDBBackend) GetModuleDownload(ctx context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error) {
	return nil, nil
}
