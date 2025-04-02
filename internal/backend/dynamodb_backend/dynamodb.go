package dynamodb_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/uuid"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/models"
	"go-terraform-registry/internal/pgp"
	registrytypes "go-terraform-registry/internal/types"
	"strings"
)

var _ backend.RegistryProviderBackend = &DynamoDBBackend{}

type DynamoDBBackend struct {
	Config config.RegistryConfig
	Tables DynamoTables

	client *dynamodb.Client
}

type DynamoTables struct {
	GPGTableName             string
	ProviderTableName        string
	ProviderVersionTableName string
	ModuleTableName          string
}

func NewDynamoDBBackend(config config.RegistryConfig) backend.RegistryProviderBackend {
	return &DynamoDBBackend{
		Config: config,
	}
}

func (d *DynamoDBBackend) ConfigureBackend(ctx context.Context) error {
	d.Tables.GPGTableName = "terraform_gpg_keys"
	d.Tables.ProviderTableName = "terraform_providers"
	d.Tables.ProviderVersionTableName = "terraform_providers_versions"
	d.Tables.ModuleTableName = "terraform_modules"

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	if d.Config.AssumeRoleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		credentials := stscreds.NewAssumeRoleProvider(stsClient, d.Config.AssumeRoleARN)
		cfg.Credentials = aws.NewCredentialsCache(credentials)
	}

	d.client = dynamodb.NewFromConfig(cfg)

	return nil
}

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

func (d *DynamoDBBackend) RegistryProviders(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProvidersRequest) (*models.RegistryProvidersResponse, error) {
	newUUID := uuid.New()
	key := fmt.Sprintf("%s:%s:%s/%s", parameters.Organization, request.Data.Attributes.RegistryName, request.Data.Attributes.Namespace, request.Data.Attributes.Name)
	p := Provider{
		ID: newUUID.String(),
	}
	err := setProvider(ctx, d.client, d.Tables.ProviderTableName, key, p)
	if err != nil {
		return nil, err
	}

	resp := &models.RegistryProvidersResponse{
		Data: models.RegistryProvidersResponseData{
			ID:   newUUID.String(),
			Type: "registry-providers",
			Attributes: models.RegistryProvidersResponseAttributes{
				Name:         request.Data.Attributes.Name,
				Namespace:    request.Data.Attributes.Namespace,
				RegistryName: request.Data.Attributes.RegistryName,
				Permissions: models.RegistryProvidersResponsePermissions{
					CanDelete: true,
				},
			},
		},
	}

	return resp, nil
}

func (d *DynamoDBBackend) GPGKey(ctx context.Context, request models.GPGKeyRequest) (*models.GPGKeyResponse, error) {
	newUUID := uuid.New()
	keyId := pgp.GetKeyID(request.Data.Attributes.AsciiArmor)

	gpg := GPGKey{
		Namespace:  request.Data.Attributes.Namespace,
		KeyID:      keyId[0],
		ID:         newUUID.String(),
		AsciiArmor: request.Data.Attributes.AsciiArmor,
	}
	err := setGPG(ctx, d.client, d.Tables.GPGTableName, gpg)
	if err != nil {
		return nil, err
	}

	resp := &models.GPGKeyResponse{
		Data: models.GPGKeyResponseData{
			ID: newUUID.String(),
			Attributes: models.GPGKeyResponseAttributes{
				AsciiArmor: request.Data.Attributes.AsciiArmor,
				KeyID:      keyId[0],
				Namespace:  request.Data.Attributes.Namespace,
			},
		},
	}

	return resp, nil
}

func (d *DynamoDBBackend) RegistryProviderVersions(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionsRequest) (*models.RegistryProviderVersionsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	provider, err := getProvider(ctx, d.client, d.Tables.ProviderTableName, key)
	if err != nil {
		return nil, err
	}

	gpg, err := getGPG(ctx, d.client, d.Tables.GPGTableName, parameters.Namespace, request.Data.Attributes.KeyID)
	if err != nil {
		return nil, err
	}
	if gpg == nil {
		return nil, fmt.Errorf("no GPG key found for %s", request.Data.Attributes.KeyID)
	}

	newUUID := uuid.New()
	pv := ProviderVersion{
		ID:            newUUID.String(),
		Version:       request.Data.Attributes.Version,
		Protocols:     request.Data.Attributes.Protocols,
		GPGKeyID:      gpg.KeyID,
		GPGASCIIArmor: gpg.AsciiArmor,
	}
	err = setProviderVersion(ctx, d.client, d.Tables.ProviderVersionTableName, *provider, pv)
	if err != nil {
		return nil, err
	}

	resp := &models.RegistryProviderVersionsResponse{
		Data: models.RegistryProviderVersionsResponseData{
			ID:   pv.ID,
			Type: "registry-provider-versions",
			Attributes: models.RegistryProviderVersionsResponseAttributes{
				Version:   pv.Version,
				Protocols: pv.Protocols,
				KeyID:     pv.GPGKeyID,
			},
		},
	}

	return resp, nil
}

func (d *DynamoDBBackend) RegistryProviderVersionPlatforms(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionPlatformsRequest) (*models.RegistryProviderVersionPlatformsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	provider, err := getProvider(ctx, d.client, d.Tables.ProviderTableName, key)
	if err != nil {
		return nil, err
	}

	pv, err := getProviderVersion(ctx, d.client, d.Tables.ProviderVersionTableName, *provider, parameters.Version)
	if err != nil {
		return nil, err
	}

	duplicate := duplicatePlatform(pv.Platform, request.Data.Attributes.OS, request.Data.Attributes.Arch)
	if duplicate {
		return nil, fmt.Errorf("duplicate platform exists matching OS and Architecture")
	}

	newUUID := uuid.New()
	platform := ProviderPlatform{
		ID:       newUUID.String(),
		OS:       request.Data.Attributes.OS,
		Arch:     request.Data.Attributes.Arch,
		SHASum:   request.Data.Attributes.Shasum,
		Filename: request.Data.Attributes.Filename,
	}

	err = appendPlatform(ctx, d.client, d.Tables.ProviderVersionTableName, provider.Provider, parameters.Version, platform)
	if err != nil {
		return nil, err
	}

	resp := &models.RegistryProviderVersionPlatformsResponse{
		Data: models.RegistryProviderVersionPlatformsResponseData{
			ID:   platform.ID,
			Type: "registry-provider-platforms",
			Attributes: models.RegistryProviderVersionPlatformsResponseAttributes{
				OS:       platform.OS,
				Arch:     platform.Arch,
				Shasum:   platform.SHASum,
				Filename: platform.Filename,
			},
		},
	}

	return resp, nil
}

func setProvider(ctx context.Context, client *dynamodb.Client, tableName string, key string, provider Provider) error {
	item := map[string]types.AttributeValue{
		"provider": &types.AttributeValueMemberS{Value: key},
		"id":       &types.AttributeValueMemberS{Value: provider.ID},
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return err
}

func getProvider(ctx context.Context, client *dynamodb.Client, tableName string, key string) (*Provider, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("provider = :p"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: key},
		},
	}

	resp, err := client.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 1 {
		p := &Provider{
			Provider: resp.Items[0]["provider"].(*types.AttributeValueMemberS).Value,
			ID:       resp.Items[0]["id"].(*types.AttributeValueMemberS).Value,
		}
		return p, nil
	}

	return nil, nil
}

func setGPG(ctx context.Context, client *dynamodb.Client, tableName string, gpg GPGKey) error {
	item := map[string]types.AttributeValue{
		"namespace":   &types.AttributeValueMemberS{Value: gpg.Namespace},
		"key_id":      &types.AttributeValueMemberS{Value: gpg.KeyID},
		"id":          &types.AttributeValueMemberS{Value: gpg.ID},
		"ascii_armor": &types.AttributeValueMemberS{Value: gpg.AsciiArmor},
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return err
}

func getGPG(ctx context.Context, client *dynamodb.Client, tableName string, namespace string, keyId string) (*GPGKey, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("namespace = :n and key_id = :k"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":n": &types.AttributeValueMemberS{Value: namespace},
			":k": &types.AttributeValueMemberS{Value: keyId},
		},
	}

	resp, err := client.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 1 {
		gpgKey := &GPGKey{
			Namespace:  namespace,
			KeyID:      keyId,
			ID:         resp.Items[0]["id"].(*types.AttributeValueMemberS).Value,
			AsciiArmor: resp.Items[0]["ascii_armor"].(*types.AttributeValueMemberS).Value,
		}
		return gpgKey, nil
	}

	return nil, nil
}

func setProviderVersion(ctx context.Context, client *dynamodb.Client, tableName string, provider Provider, providerVersion ProviderVersion) error {
	item := map[string]types.AttributeValue{
		"provider":        &types.AttributeValueMemberS{Value: provider.Provider},
		"version":         &types.AttributeValueMemberS{Value: providerVersion.Version},
		"id":              &types.AttributeValueMemberS{Value: providerVersion.ID},
		"gpg_key_id":      &types.AttributeValueMemberS{Value: providerVersion.GPGKeyID},
		"gpg_ascii_armor": &types.AttributeValueMemberS{Value: providerVersion.GPGASCIIArmor},
		"platforms":       &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		"protocols":       &types.AttributeValueMemberSS{Value: providerVersion.Protocols},
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return err
}

func getProviderVersion(ctx context.Context, client *dynamodb.Client, tableName string, provider Provider, version string) (*ProviderVersion, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("provider = :p and version = :v"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: provider.Provider},
			":v": &types.AttributeValueMemberS{Value: version},
		},
	}

	resp, err := client.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query items, %v", err)
	}

	if resp.Count == 1 {
		pv := ProviderVersion{
			ID:            resp.Items[0]["id"].(*types.AttributeValueMemberS).Value,
			Version:       resp.Items[0]["version"].(*types.AttributeValueMemberS).Value,
			Protocols:     resp.Items[0]["protocols"].(*types.AttributeValueMemberSS).Value,
			GPGKeyID:      resp.Items[0]["gpg_key_id"].(*types.AttributeValueMemberS).Value,
			GPGASCIIArmor: resp.Items[0]["gpg_ascii_armor"].(*types.AttributeValueMemberS).Value,
		}

		platformsList := resp.Items[0]["platforms"].(*types.AttributeValueMemberL)
		for _, attr := range platformsList.Value {
			strAttr := attr.(*types.AttributeValueMemberS)

			var platform ProviderPlatform
			err := json.Unmarshal([]byte(strAttr.Value), &platform)
			if err != nil {
				return nil, err
			}

			pv.Platform = append(pv.Platform, platform)
		}

		return &pv, nil
	}

	return nil, nil
}

func duplicatePlatform(platforms []ProviderPlatform, os string, arch string) bool {
	for _, platform := range platforms {
		if strings.EqualFold(platform.OS, os) && strings.EqualFold(platform.Arch, arch) {
			return true
		}
	}

	return false
}

func appendPlatform(ctx context.Context, client *dynamodb.Client, tableName, provider string, version string, platform ProviderPlatform) error {
	platformJSON, err := json.Marshal(platform)
	if err != nil {
		return fmt.Errorf("failed to marshal platform: %w", err)
	}

	params := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"provider": &types.AttributeValueMemberS{Value: provider},
			"version":  &types.AttributeValueMemberS{Value: version},
		},
		UpdateExpression: aws.String("SET platforms = list_append(platforms, :platform)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":platform": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: string(platformJSON)},
				},
			},
		},
	}

	_, err = client.UpdateItem(ctx, params)

	return err
}
