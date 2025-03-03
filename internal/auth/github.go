package auth

import (
	"context"
	"fmt"
	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

func GetGitHubUserName(ctx context.Context, token string) (*string, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error getting user: %s", err)
	}

	return user.Login, nil
}
