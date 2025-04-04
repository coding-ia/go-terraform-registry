package config

import (
	"os"
	"strings"
)

type RegistryConfig struct {
	AllowAnonymousAccess   bool
	AssumeRoleARN          string
	Backend                string
	GitHubEndpoint         string
	OauthClientID          string
	OauthClientRedirectURL string
	OauthClientSecret      string
	Organization           string
	S3BucketName           string
	S3BucketRegion         string
	StorageBackend         string
	TokenEncryptionKey     string
}

func GetRegistryConfig() RegistryConfig {
	config := RegistryConfig{
		AllowAnonymousAccess:   getBoolEnv("ALLOW_ANONYMOUS_ACCESS", true),
		AssumeRoleARN:          os.Getenv("ASSUME_ROLE_ARN"),
		Backend:                os.Getenv("BACKEND"),
		GitHubEndpoint:         os.Getenv("GITHUB_ENDPOINT"),
		OauthClientID:          os.Getenv("OAUTH_CLIENT_ID"),
		OauthClientRedirectURL: os.Getenv("OAUTH_CLIENT_REDIRECT_URL"),
		OauthClientSecret:      os.Getenv("OAUTH_CLIENT_SECRET"),
		Organization:           os.Getenv("DEFAULT_ORGANIZATION"),
		S3BucketName:           os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion:         os.Getenv("S3_BUCKET_REGION"),
		StorageBackend:         os.Getenv("STORAGE_BACKEND"),
		TokenEncryptionKey:     os.Getenv("TOKEN_ENCRYPTION_KEY"),
	}
	if config.Organization == "" {
		config.Organization = "default"
	}
	return config
}

func getBoolEnv(key string, defaultValue bool) bool {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	val = strings.ToLower(val)
	return val == "true" || val == "1"
}
