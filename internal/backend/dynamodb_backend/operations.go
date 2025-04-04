package dynamodb_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strings"
)

func setProvider(ctx context.Context, client *dynamodb.Client, tableName string, key string, provider Provider) error {
	item := map[string]types.AttributeValue{
		"provider": &types.AttributeValueMemberS{Value: key},
		"id":       &types.AttributeValueMemberS{Value: provider.ID},
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(provider)"),
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
		TableName:           aws.String(tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(provider) and attribute_not_exists(version)"),
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
