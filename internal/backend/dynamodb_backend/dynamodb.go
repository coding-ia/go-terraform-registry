package dynamodb_backend

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"strings"
)

var _ backend.RegistryProviderBackend = &DynamoDBBackend{}

type DynamoDBBackend struct {
	ProviderTableName string
	ModuleTableName   string
}

func NewDynamoDBBackend() backend.RegistryProviderBackend {
	return &DynamoDBBackend{}
}

func (d *DynamoDBBackend) ConfigureBackend(_ context.Context) {
	d.ProviderTableName = "terraform_providers"
	d.ModuleTableName = "terraform_modules"
}

func (d *DynamoDBBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters) (*models.TerraformProviderPlatformResponse, error) {
	providerName := fmt.Sprintf("%s/%s", parameters.Namespace, parameters.Name)

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.ProviderTableName),
		KeyConditionExpression: aws.String("provider = :p and #v = :v"),
		ExpressionAttributeNames: map[string]string{
			"#v": "version",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: providerName},
			":v": &types.AttributeValueMemberS{Value: parameters.Version},
		},
	}

	svc, err := createDynamoDBClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 1 {
		protocols := resp.Items[0]["protocols"].(*types.AttributeValueMemberSS).Value
		shaSumsUrl := resp.Items[0]["shasums_url"].(*types.AttributeValueMemberS).Value
		shaSumsSignatureUrl := resp.Items[0]["shasums_signature_url"].(*types.AttributeValueMemberS).Value

		var keys []models.GPGPublicKeys
		gpgKeys := resp.Items[0]["gpg_public_keys"].(*types.AttributeValueMemberL).Value
		for _, item := range gpgKeys {
			m := item.(*types.AttributeValueMemberM).Value
			keyId := m["key_id"].(*types.AttributeValueMemberS).Value
			asciiArmor := m["ascii_armor"].(*types.AttributeValueMemberS).Value

			keys = append(keys, models.GPGPublicKeys{
				KeyId:      keyId,
				AsciiArmor: asciiArmor,
			})
		}

		response := &models.TerraformProviderPlatformResponse{
			Protocols:           protocols,
			ShasumsUrl:          shaSumsUrl,
			ShasumsSignatureUrl: shaSumsSignatureUrl,
			SigningKeys: models.SigningKeys{
				GPGPublicKeys: keys,
			},
		}

		releaseList := resp.Items[0]["release"].(*types.AttributeValueMemberL).Value
		for _, item := range releaseList {
			releaseItem := item.(*types.AttributeValueMemberM).Value
			os := releaseItem["os"].(*types.AttributeValueMemberS).Value
			arch := releaseItem["arch"].(*types.AttributeValueMemberS).Value

			if strings.EqualFold(os, parameters.OS) &&
				strings.EqualFold(arch, parameters.Architecture) {
				response.Filename = releaseItem["filename"].(*types.AttributeValueMemberS).Value
				response.DownloadUrl = releaseItem["download_url"].(*types.AttributeValueMemberS).Value
				response.Shasum = releaseItem["shasum"].(*types.AttributeValueMemberS).Value
				response.OS = os
				response.Arch = arch

				return response, nil
			}
		}
	}

	return nil, nil
}

func (d *DynamoDBBackend) GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters) (*models.TerraformAvailableProvider, error) {
	providerName := fmt.Sprintf("%s/%s", parameters.Namespace, parameters.Name)

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.ProviderTableName),
		KeyConditionExpression: aws.String("provider = :p"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: providerName},
		},
	}

	svc, err := createDynamoDBClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Query(ctx, params)
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

		releaseList := item["release"].(*types.AttributeValueMemberL)
		for _, release := range releaseList.Value {
			releaseItem := release.(*types.AttributeValueMemberM)

			v.Platforms = append(v.Platforms, models.TerraformAvailablePlatform{
				OS:   extractString(releaseItem.Value, "os"),
				Arch: extractString(releaseItem.Value, "arch"),
			})
		}
		versions = append(versions, v)
	}

	provider := &models.TerraformAvailableProvider{
		Versions: versions,
	}

	return provider, nil
}

func (d *DynamoDBBackend) GetModuleVersions(ctx context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error) {
	moduleName := fmt.Sprintf("%s/%s/%s", parameters.Namespace, parameters.Name, parameters.System)

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.ModuleTableName),
		KeyConditionExpression: aws.String("#m = :m"),
		ExpressionAttributeNames: map[string]string{
			"#m": "module",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":m": &types.AttributeValueMemberS{Value: moduleName},
		},
	}

	svc, err := createDynamoDBClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 0 {
		return nil, nil
	}

	var versions []models.TerraformAvailableModuleVersion
	if resp.Count > 0 {
		for _, item := range resp.Items {
			version := item["version"].(*types.AttributeValueMemberS).Value
			versions = append(versions, models.TerraformAvailableModuleVersion{
				Version: version,
			})
		}
	}

	modules := &models.TerraformAvailableModule{
		Modules: []models.TerraformAvailableModuleVersions{
			{
				Versions: versions,
			},
		},
	}

	return modules, nil
}

func (d *DynamoDBBackend) GetModuleDownload(ctx context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error) {
	moduleName := fmt.Sprintf("%s/%s/%s", parameters.Namespace, parameters.Name, parameters.System)

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.ModuleTableName),
		KeyConditionExpression: aws.String("#m = :m and version = :v"),
		ExpressionAttributeNames: map[string]string{
			"#m": "module",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":m": &types.AttributeValueMemberS{Value: moduleName},
			":v": &types.AttributeValueMemberS{Value: parameters.Version},
		},
	}

	svc, err := createDynamoDBClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 1 {
		uri := resp.Items[0]["download_url"].(*types.AttributeValueMemberS).Value
		return &uri, nil
	}

	return nil, nil
}

func (d *DynamoDBBackend) RegistryProviders(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProvidersRequest) (*models.RegistryProvidersResponse, error) {
	return nil, nil
}

func (d *DynamoDBBackend) GPGKey(ctx context.Context, request models.GPGKeyRequest) (*models.GPGKeyResponse, error) {
	return nil, nil
}

func (d *DynamoDBBackend) RegistryProviderVersions(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionsRequest) (*models.RegistryProviderVersionsResponse, error) {
	return nil, nil
}

func (d *DynamoDBBackend) RegistryProviderVersionPlatforms(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionPlatformsRequest) (*models.RegistryProviderVersionPlatformsResponse, error) {
	return nil, nil
}

func extractString(m map[string]types.AttributeValue, key string) string {
	if v, ok := m[key].(*types.AttributeValueMemberS); ok {
		return v.Value
	}
	return ""
}

func createDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}
