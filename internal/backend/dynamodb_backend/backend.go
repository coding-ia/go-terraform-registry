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

var _ backend.BackendLifecycle = &DynamoDBBackend{}

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

	return &backend.Backend{
		RegistryBackend:         b,
		ProvidersBackend:        b,
		ProviderVersionsBackend: b,
		ModulesBackend:          b,
		GPGKeysBackend:          b,
	}, nil
}

func (d *DynamoDBBackend) Configure(ctx context.Context) error {
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

	log.Println("Using DynamoDB for backend.")

	return nil
}

func (d *DynamoDBBackend) Close(ctx context.Context) error {
	return nil
}
