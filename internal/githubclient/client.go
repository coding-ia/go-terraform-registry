package githubclient

import (
	"context"
	"fmt"
	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
	"net/http"
)

func NewClient(ctx context.Context, token, githubBaseURL string) (*github.Client, error) {
	var httpClient *http.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(ctx, ts)
	} else {
		httpClient = http.DefaultClient
	}

	if githubBaseURL == "" {
		return github.NewClient(httpClient), nil
	}

	baseUrl := fmt.Sprintf("%sapi/v3/", githubBaseURL)
	uploadUrl := fmt.Sprintf("%sapi/uploads/", githubBaseURL)
	return github.NewClient(httpClient).WithEnterpriseURLs(baseUrl, uploadUrl)
}
