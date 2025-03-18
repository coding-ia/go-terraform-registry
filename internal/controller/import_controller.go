package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"golang.org/x/crypto/openpgp"
	"io"
	"log"
	"net/http"
	"strings"
)

type ImportController struct {
	Backend backend.RegistryProviderBackend
}

type RegistryImportController interface {
	ProviderImport(*gin.Context)
	ModuleImport(*gin.Context)
}

func NewImportController(r *gin.Engine, backend backend.RegistryProviderBackend) RegistryImportController {
	ic := &ImportController{
		Backend: backend,
	}

	importEndpoint := r.Group("/import")

	importEndpoint.POST("/provider", ic.ProviderImport)
	importEndpoint.POST("/module", ic.ModuleImport)

	return ic
}

func (i *ImportController) ProviderImport(c *gin.Context) {
	var requestData models.ImportProviderData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetReleaseByTag(c.Request.Context(), requestData.Owner, requestData.Repository, requestData.Tag)
	if err != nil {
		fmt.Printf("Error getting release: %v", err)
	}

	fmt.Printf("Release: %s\n", release.GetName())

	var shas map[string]string
	var shaUrl string
	var shaSigUrl string
	var manifest models.ProviderManifest
	providers := make(map[string]string)

	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.GetName(), "_SHA256SUMS") {
			shaUrl = asset.GetBrowserDownloadURL()
			content, err := getHTTPContent(shaUrl)
			if err != nil {
				fmt.Printf("Error getting SHA256SUMS: %v", err)
				return
			}
			shas, err = parseSha256SUMS(content)
			if err != nil {
				fmt.Printf("Error parsing SHA256SUMS: %v", err)
				return
			}
			continue
		}

		if strings.HasSuffix(asset.GetName(), "_manifest.json") {
			content, err := getHTTPContent(asset.GetBrowserDownloadURL())
			if err != nil {
				fmt.Printf("Error getting manifest: %v", err)
				return
			}
			err = json.Unmarshal(content, &manifest)
			if err != nil {
				fmt.Printf("Error parsing manifest: %v", err)
				return
			}
			continue
		}

		if strings.HasSuffix(asset.GetName(), "SHA256SUMS.sig") {
			shaSigUrl = asset.GetBrowserDownloadURL()
			continue
		}

		providers[asset.GetName()] = asset.GetBrowserDownloadURL()
	}

	fingerprint := getKeyFingerprint(requestData.GPGPublicKey)

	provider := &registrytypes.ProviderImport{
		Name:           requestData.Name,
		SHASUMUrl:      shaUrl,
		SHASUMSigUrl:   shaSigUrl,
		Protocols:      manifest.Metadata.ProtocolVersions,
		GPGASCIIArmor:  requestData.GPGPublicKey,
		GPGFingerprint: fingerprint[0],
	}

	for k, v := range providers {
		name := strings.TrimSuffix(k, ".zip")
		parts := strings.Split(name, "_")

		provider.Version = parts[1]
		pri := &registrytypes.ProviderReleaseImport{
			DownloadUrl:  v,
			Filename:     k,
			SHASUM:       shas[k],
			OS:           parts[2],
			Architecture: parts[3],
		}

		provider.Release = append(provider.Release, *pri)
	}

	err = i.Backend.ImportProvider(c.Request.Context(), *provider)
	if err != nil {
		fmt.Printf("Error creating provider entry: %v", err)
	}
}

func (i *ImportController) ModuleImport(_ *gin.Context) {

}

func parseSha256SUMS(content []byte) (map[string]string, error) {
	dataMap := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) != 2 {
			continue
		}

		key := fields[1]
		value := fields[0]
		dataMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dataMap, nil
}

func getHTTPContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getKeyFingerprint(publicKey string) []string {
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(publicKey))
	if err != nil {
		log.Fatal(err)
	}

	var keys []string
	for _, entity := range entityList {
		fingerPrint := entity.PrimaryKey.Fingerprint
		value := fmt.Sprintf("Key ID: %x\n", fingerPrint)
		keys = append(keys, strings.ToUpper(value))
	}

	return keys
}
