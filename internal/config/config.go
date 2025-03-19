package config

import (
	"os"
	"strings"
)

type RegistryConfig struct {
	AllowAnonymousAccess   bool
	Backend                string
	S3BucketName           string
	S3BucketRegion         string
	TokenEncryptionKey     string
	OauthAuthURL           string
	OauthClientID          string
	OauthClientRedirectURL string
	OauthClientSecret      string
	OauthTokenURL          string
}

func GetRegistryConfig() RegistryConfig {
	config := RegistryConfig{
		AllowAnonymousAccess:   getBoolEnv("ALLOW_ANONYMOUS_ACCESS", true),
		Backend:                os.Getenv("BACKEND"),
		S3BucketName:           os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion:         os.Getenv("S3_BUCKET_REGION"),
		TokenEncryptionKey:     os.Getenv("TOKEN_ENCRYPTION_KEY"),
		OauthAuthURL:           os.Getenv("OAUTH_CLIENT_AUTH_URL"),
		OauthClientID:          os.Getenv("OAUTH_CLIENT_ID"),
		OauthClientRedirectURL: os.Getenv("OAUTH_CLIENT_REDIRECT_URL"),
		OauthClientSecret:      os.Getenv("OAUTH_CLIENT_SECRET"),
		OauthTokenURL:          os.Getenv("OAUTH_CLIENT_TOKEN_URL"),
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
