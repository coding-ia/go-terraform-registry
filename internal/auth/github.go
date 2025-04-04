package auth

import (
	"context"
	"fmt"
	"github.com/google/go-github/v69/github"
)

func GetGitHubUserName(ctx context.Context, client *github.Client, token string) (*string, error) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error getting user: %s", err)
	}

	return user.Login, nil
}
