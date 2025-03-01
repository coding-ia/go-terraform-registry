package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/models"
	"io"
	"log"
	"net/http"
	"strings"
)

type ProviderController struct {
	Config registryconfig.RegistryConfig
}

type RegistryProviderController interface {
	ProviderPackage(*gin.Context)
	Versions(*gin.Context)
}

func NewProviderController(r *gin.Engine, config registryconfig.RegistryConfig) RegistryProviderController {
	pc := &ProviderController{
		Config: config,
	}

	providers := r.Group("/terraform/providers/v1")
	{
		providers.GET("/:ns/:name/versions", pc.Versions)
		providers.GET("/:ns/:name/:version/download/:os/:arch", pc.ProviderPackage)
	}

	return pc
}

func (p *ProviderController) ProviderPackage(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	version := c.Param("version")
	os := c.Param("os")
	arch := c.Param("arch")

	path := fmt.Sprintf("providers/%s/%s/provider.json", ns, name)
	providerData, err := getS3ProviderData(c.Request.Context(), p.Config.S3BucketName, path)
	if err != nil {
		log.Printf(err.Error())
		errorResponse(c)
		return
	}

	if providerData != nil {
		provider := matchProviderVersion(providerData.Versions, version)

		if provider != nil {
			platform := matchProviderPlatform(provider.Platforms, os, arch)
			platform.Protocols = provider.Protocols
			c.JSON(http.StatusOK, platform)
			return
		}
	}

	errorResponseWithMessage(c, "No provider found")
}

func (p *ProviderController) Versions(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")

	path := fmt.Sprintf("providers/%s/%s/provider.json", ns, name)
	providerData, err := getS3ProviderData(c.Request.Context(), p.Config.S3BucketName, path)
	if err != nil {
		log.Printf(err.Error())
		errorResponse(c)
		return
	}

	if providerData != nil {
		var providers models.TerraformAvailableProvider
		for _, provider := range providerData.Versions {
			version := models.TerraformAvailableVersion{
				Version:   provider.Version,
				Protocols: provider.Protocols,
			}

			for _, platform := range provider.Platforms {
				p := models.TerraformAvailablePlatform{
					OS:   platform.OS,
					Arch: platform.Arch,
				}
				version.Platforms = append(version.Platforms, p)
			}

			providers.Versions = append(providers.Versions, version)
		}

		c.JSON(http.StatusOK, providers)
		return
	}

	errorResponseWithMessage(c, "No provider found")
}

func getS3ProviderData(ctx context.Context, bucket, key string) (*models.TerraformProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	getObjectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get object from S3, %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("unable to close S3, %v", err)
		}
	}(getObjectOutput.Body)

	// Read the file content into memory
	fileContent, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read object body, %v", err)
	}

	var data models.TerraformProvider
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON, %v", err)
	}

	return &data, nil
}

func matchProviderVersion(providerVersions []models.TerraformProviderVersion, version string) *models.TerraformProviderVersion {
	for _, provider := range providerVersions {
		if strings.EqualFold(version, provider.Version) {
			return &provider
		}
	}

	return nil
}

func matchProviderPlatform(platforms []models.TerraformProviderPlatform, os string, arch string) *models.TerraformProviderPlatformResponse {
	for _, platform := range platforms {
		if strings.EqualFold(os, platform.OS) &&
			strings.EqualFold(arch, platform.Arch) {
			response := models.TerraformProviderPlatformResponse{
				OS:                  platform.OS,
				Arch:                platform.Arch,
				Filename:            platform.Filename,
				DownloadUrl:         platform.DownloadUrl,
				ShasumsUrl:          platform.ShasumsUrl,
				ShasumsSignatureUrl: platform.ShasumsSignatureUrl,
				Shasum:              platform.Shasum,
				SigningKeys:         platform.SigningKeys,
			}

			return &response
		}
	}

	return nil
}
