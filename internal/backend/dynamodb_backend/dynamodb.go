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
	"log"
	"strings"
)

var _ backend.RegistryProviderBackend = &DynamoDBBackend{}

type DynamoDBBackend struct {
	TableName string
}

func NewDynamoDBBackend() backend.RegistryProviderBackend {
	return &DynamoDBBackend{}
}

func (d *DynamoDBBackend) ConfigureBackend(_ context.Context) {
	d.TableName = "terraform_providers"
}

func (d *DynamoDBBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters) (*models.TerraformProviderPlatformResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	providerName := fmt.Sprintf("%s/%s", parameters.Namespace, parameters.Name)
	releaseName := fmt.Sprintf("%s#%s#%s", parameters.Version, parameters.OS, parameters.Architecture)

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.TableName),
		KeyConditionExpression: aws.String("provider = :p and #r = :r"),
		ExpressionAttributeNames: map[string]string{
			"#r": "release",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: providerName},
			":r": &types.AttributeValueMemberS{Value: releaseName},
		},
	}

	resp, err := svc.Query(ctx, params)
	if err != nil {
		log.Fatalf("failed to query items, %v", err)
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
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	providerName := fmt.Sprintf("%s/%s", parameters.Namespace, parameters.Name)

	params := &dynamodb.QueryInput{
		TableName:              aws.String(d.TableName),
		KeyConditionExpression: aws.String("provider = :p"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: providerName},
		},
	}

	resp, err := svc.Query(ctx, params)
	if err != nil {
		log.Fatalf("failed to query items, %v", err)
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
