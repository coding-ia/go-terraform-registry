package config

import "os"

type RegistryConfig struct {
	S3BucketName   string
	S3BucketRegion string
}

func GetRegistryConfig() RegistryConfig {
	config := RegistryConfig{
		S3BucketName:   os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion: os.Getenv("S3_BUCKET_REGION"),
	}
	return config
}
