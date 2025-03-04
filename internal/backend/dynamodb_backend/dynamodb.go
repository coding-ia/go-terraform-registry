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
	releaseName := fmt.Sprintf("%s#%s#%s", parameters.Version, parameters.OS, parameters.Architecture)

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.ProviderTableName),
		KeyConditionExpression: aws.String("provider = :p and #r = :r"),
		ExpressionAttributeNames: map[string]string{
			"#r": "release",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: providerName},
			":r": &types.AttributeValueMemberS{Value: releaseName},
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
		release := resp.Items[0]["release"].(*types.AttributeValueMemberS).Value
		filename := resp.Items[0]["filename"].(*types.AttributeValueMemberS).Value
		downloadUrl := resp.Items[0]["download_url"].(*types.AttributeValueMemberS).Value
		shaSumsUrl := resp.Items[0]["shasums_url"].(*types.AttributeValueMemberS).Value
		shaSumsSignatureUrl := resp.Items[0]["shasums_signature_url"].(*types.AttributeValueMemberS).Value
		shaSum := resp.Items[0]["shasum"].(*types.AttributeValueMemberS).Value
		protocols := resp.Items[0]["protocols"].(*types.AttributeValueMemberSS).Value
		gpgKeys := resp.Items[0]["gpg_public_keys"].(*types.AttributeValueMemberL).Value

		var keys []models.GPGPublicKeys
		for _, item := range gpgKeys {
			m := item.(*types.AttributeValueMemberM).Value
			keyId := m["key_id"].(*types.AttributeValueMemberS).Value
			asciiArmor := m["ascii_armor"].(*types.AttributeValueMemberS).Value

			keys = append(keys, models.GPGPublicKeys{
				KeyId:      keyId,
				AsciiArmor: asciiArmor,
			})
		}

		parts := strings.Split(release, "#")

		response := &models.TerraformProviderPlatformResponse{
			Protocols:           protocols,
			OS:                  parts[1],
			Arch:                parts[2],
			Filename:            filename,
			DownloadUrl:         downloadUrl,
			ShasumsUrl:          shaSumsUrl,
			ShasumsSignatureUrl: shaSumsSignatureUrl,
			Shasum:              shaSum,
			SigningKeys: models.SigningKeys{
				GPGPublicKeys: keys,
			},
		}

		return response, nil
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

	versions := make(map[string]models.TerraformAvailableVersion)
	for _, item := range resp.Items {
		release := item["release"].(*types.AttributeValueMemberS).Value
		protocols := item["protocols"].(*types.AttributeValueMemberSS).Value
		parts := strings.Split(release, "#")

		value, ok := versions[parts[0]]

		if ok {
			value.Platforms = append(value.Platforms, models.TerraformAvailablePlatform{
				OS:   parts[1],
				Arch: parts[2],
			})
			mergeProtocols(&value.Protocols, protocols)
			versions[parts[0]] = value
		} else {
			versions[parts[0]] = models.TerraformAvailableVersion{
				Version:   parts[0],
				Protocols: protocols,
				Platforms: []models.TerraformAvailablePlatform{
					{
						OS:   parts[1],
						Arch: parts[2],
					},
				},
			}
		}
	}

	provider := &models.TerraformAvailableProvider{}
	if versions != nil {
		for _, value := range versions {
			provider.Versions = append(provider.Versions, value)
		}
	}

	return provider, nil
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

func createDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func mergeProtocols(protocols *[]string, additionalProtocols []string) {
	uniqueSet := make(map[string]struct{}, len(*protocols))
	var result []string

	for _, protocol := range *protocols {
		uniqueSet[protocol] = struct{}{}
		result = append(result, protocol)
	}

	for _, item := range additionalProtocols {
		if _, exists := uniqueSet[item]; !exists {
			uniqueSet[item] = struct{}{}
			result = append(result, item)
		}
	}

	*protocols = result
}
