package dynamodb_backend

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/models"
	"go-terraform-registry/internal/pgp"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.RegistryProviderBackend = &DynamoDBBackend{}

type DynamoDBBackend struct {
	Config config.RegistryConfig
	Tables DynamoTables
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

func (d *DynamoDBBackend) ConfigureBackend(ctx context.Context) {
	d.Tables.GPGTableName = "terraform_gpg_keys"
	d.Tables.ProviderTableName = "terraform_providers"
	d.Tables.ProviderVersionTableName = "terraform_providers_versions"
	d.Tables.ModuleTableName = "terraform_modules"
}

func (d *DynamoDBBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters, userParameters registrytypes.UserParameters) (*models.TerraformProviderPlatformResponse, error) {
	return nil, nil
}

func (d *DynamoDBBackend) GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters, userParameters registrytypes.UserParameters) (*models.TerraformAvailableProvider, error) {
	return nil, nil
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
	err := setProvider(ctx, d.Tables.ProviderTableName, key, p)
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
	err := setGPG(ctx, d.Tables.GPGTableName, gpg)
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
	provider, err := getProvider(ctx, d.Tables.ProviderTableName, key)
	if err != nil {
		return nil, err
	}

	gpg, err := getGPG(ctx, d.Tables.GPGTableName, parameters.Namespace, request.Data.Attributes.KeyID)
	if err != nil {
		return nil, err
	}

	newUUID := uuid.New()
	pv := ProviderVersion{
		ID:            newUUID.String(),
		Version:       request.Data.Attributes.Version,
		Protocols:     request.Data.Attributes.Protocols,
		GPGKeyID:      gpg.KeyID,
		GPGASCIIArmor: gpg.AsciiArmor,
	}
	err = setProviderVersion(ctx, d.Tables.ProviderVersionTableName, *provider, pv)
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
	return nil, nil
}

func extractString(m map[string]types.AttributeValue, key string) string {
	if v, ok := m[key].(*types.AttributeValueMemberS); ok {
		return v.Value
	}
	return ""
}

func createDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func setProvider(ctx context.Context, tableName string, key string, provider Provider) error {
	item := map[string]types.AttributeValue{
		"provider": &types.AttributeValueMemberS{Value: key},
		"id":       &types.AttributeValueMemberS{Value: provider.ID},
	}

	svc, err := createDynamoDBClient(ctx)
	if err != nil {
		return err
	}
	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return nil
}

func getProvider(ctx context.Context, tableName string, key string) (*Provider, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("provider = :p"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: key},
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
		p := &Provider{
			ID: resp.Items[0]["id"].(*types.AttributeValueMemberS).Value,
		}
		return p, nil
	}

	return nil, nil
}

func setGPG(ctx context.Context, tableName string, gpg GPGKey) error {
	item := map[string]types.AttributeValue{
		"namespace":   &types.AttributeValueMemberS{Value: gpg.Namespace},
		"key_id":      &types.AttributeValueMemberS{Value: gpg.KeyID},
		"id":          &types.AttributeValueMemberS{Value: gpg.ID},
		"ascii_armor": &types.AttributeValueMemberS{Value: gpg.AsciiArmor},
	}

	svc, err := createDynamoDBClient(ctx)
	if err != nil {
		return err
	}
	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return nil
}

func getGPG(ctx context.Context, tableName string, namespace string, keyId string) (*GPGKey, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("namespace = :n and key_id = :k"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":n": &types.AttributeValueMemberS{Value: namespace},
			":k": &types.AttributeValueMemberS{Value: keyId},
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

func setProviderVersion(ctx context.Context, tableName string, provider Provider, providerVersion ProviderVersion) error {
	item := map[string]types.AttributeValue{
		"provider_id":     &types.AttributeValueMemberS{Value: provider.ID},
		"version":         &types.AttributeValueMemberS{Value: providerVersion.Version},
		"protocols":       &types.AttributeValueMemberSS{Value: providerVersion.Protocols},
		"gpg_key_id":      &types.AttributeValueMemberS{Value: providerVersion.GPGKeyID},
		"gpg_ascii_armor": &types.AttributeValueMemberS{Value: providerVersion.GPGASCIIArmor},
	}

	svc, err := createDynamoDBClient(ctx)
	if err != nil {
		return err
	}
	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return nil
}
