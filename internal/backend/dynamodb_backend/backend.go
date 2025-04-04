package dynamodb_backend

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"log"
)

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

func NewDynamoDBBackend(ctx context.Context, config config.RegistryConfig) (*backend.Backend, error) {
	b := &DynamoDBBackend{
		Config: config,
	}

	err := configureBackend(ctx, b)
	if err != nil {
		return nil, err
	}

	return &backend.Backend{
		RegistryBackend:         b,
		ProvidersBackend:        b,
		ProviderVersionsBackend: b,
		GPGKeysBackend:          b,
	}, nil
}

func configureBackend(ctx context.Context, dynamoDBBackend *DynamoDBBackend) error {
	dynamoDBBackend.Tables.GPGTableName = "terraform_gpg_keys"
	dynamoDBBackend.Tables.ProviderTableName = "terraform_providers"
	dynamoDBBackend.Tables.ProviderVersionTableName = "terraform_providers_versions"
	dynamoDBBackend.Tables.ModuleTableName = "terraform_modules"

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	if dynamoDBBackend.Config.AssumeRoleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		credentials := stscreds.NewAssumeRoleProvider(stsClient, dynamoDBBackend.Config.AssumeRoleARN)
		cfg.Credentials = aws.NewCredentialsCache(credentials)
	}

	dynamoDBBackend.client = dynamodb.NewFromConfig(cfg)

	log.Println("Using DynamoDB for backend.")

	return nil
}
