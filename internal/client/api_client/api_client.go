package api_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type APIClient struct {
	Token  string
	Client *http.Client
}

func NewAPIClient(token string) *APIClient {
	return &APIClient{
		Token:  token,
		Client: &http.Client{},
	}
}

func (c *APIClient) SendRequest(method, url string, requestBody any, result any) error {
	var bodyReader io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("error marshalling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, respBody)
	}

	if result != nil {
		err = json.Unmarshal(respBody, result)
		if err != nil {
			return fmt.Errorf("error unmarshalling response: %w", err)
		}
	}

	return nil
}
