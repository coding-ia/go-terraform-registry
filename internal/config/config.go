package config

import (
	"os"
	"strings"
)

type RegistryConfig struct {
	AllowAnonymousAccess bool
	S3BucketName         string
	S3BucketRegion       string
	TokenEncryptionKey   string
	OauthClientID        string
	OauthClientSecret    string
}

func GetRegistryConfig() RegistryConfig {
	config := RegistryConfig{
		AllowAnonymousAccess: getBoolEnv("ALLOW_ANONYMOUS_ACCESS", true),
		S3BucketName:         os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion:       os.Getenv("S3_BUCKET_REGION"),
		TokenEncryptionKey:   os.Getenv("TOKEN_ENCRYPTION_KEY"),
		OauthClientID:        os.Getenv("OAUTH_CLIENT_ID"),
		OauthClientSecret:    os.Getenv("OAUTH_CLIENT_SECRET"),
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
