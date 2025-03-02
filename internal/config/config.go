package config

import "os"

type RegistryConfig struct {
	S3BucketName      string
	S3BucketRegion    string
	OauthClientID     string
	OauthClientSecret string
}

func GetRegistryConfig() RegistryConfig {
	config := RegistryConfig{
		S3BucketName:      os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion:    os.Getenv("S3_BUCKET_REGION"),
		OauthClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		OauthClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
	}
	return config
}
