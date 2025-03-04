package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v69/github"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/oauth2"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type PublishOptions struct {
	GitHubOwner  string
	GitHubRepo   string
	GitHubTag    string
	ProviderName string
	GPGPublicKey string
}

var publishOptions = &PublishOptions{}

var publishProviderCmd = &cobra.Command{
	Use:   "publish-provider",
	Short: "Publish provider to registry",
	Run: func(cmd *cobra.Command, args []string) {
		publishProvider(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(publishProviderCmd)

	publishProviderCmd.Flags().StringVar(&publishOptions.GitHubOwner, "owner", "", "GitHub owner.")
	publishProviderCmd.Flags().StringVar(&publishOptions.GitHubRepo, "repo", "", "GitHub repo.")
	publishProviderCmd.Flags().StringVar(&publishOptions.GitHubTag, "tag", "", "GitHub tag.")
	publishProviderCmd.Flags().StringVar(&publishOptions.ProviderName, "name", "", "Provider name.")
	publishProviderCmd.Flags().StringVar(&publishOptions.GPGPublicKey, "public-key", "", "GPG public key.")
}

func publishProvider(ctx context.Context) {
	owner := publishOptions.GitHubOwner
	repo := publishOptions.GitHubRepo
	tag := publishOptions.GitHubTag

	if !strings.Contains(publishOptions.ProviderName, "/") {
		fmt.Printf("Provider name must contain the namspace and name sperated by \"/\".")
		return
	}

	token := os.Getenv("GITHUB_TOKEN") // Set this in your environment
	var client *github.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
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

	asciiArmor, err := readFileContents(publishOptions.GPGPublicKey)
	if err != nil {
		fmt.Printf("Error reading GPG public key: %v", err)
		return
	}
	fingerprint := getKeyFingerprint(asciiArmor)

	provider := &registrytypes.ProviderImport{
		Name:           publishOptions.ProviderName,
		SHASUMUrl:      shaUrl,
		SHASUMSigUrl:   shaSigUrl,
		Protocols:      manifest.Metadata.ProtocolVersions,
		GPGASCIIArmor:  asciiArmor,
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

	b := config.SelectBackend(ctx, "dynamodb")
	err = b.ImportProvider(ctx, *provider)
	if err != nil {
		fmt.Printf("Error creating provider entry: %v", err)
	}
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

func readFileContents(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
