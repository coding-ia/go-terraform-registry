package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/models"
	"io"
	"net/http"
	"os"
)

type PublishOptions struct {
	Endpoint     string
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

	publishProviderCmd.Flags().StringVar(&publishOptions.Endpoint, "endpoint", "", "registry endpoint")
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

	asciiArmor, err := readFileContents(publishOptions.GPGPublicKey)
	if err != nil {
		fmt.Printf("Error reading GPG public key: %v", err)
		return
	}

	payload := models.ImportProviderData{
		Owner:        owner,
		Repository:   repo,
		Tag:          tag,
		Name:         publishOptions.ProviderName,
		GPGPublicKey: asciiArmor,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("error marshalling data: %v", err)
	}

	fmt.Println("JSON Payload:")
	fmt.Println(string(jsonData))

	resp, err := http.Post(publishOptions.Endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("error publishing provider: %v", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error: %v", err)
		}
	}(resp.Body)
}

func readFileContents(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

/*
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

*/
