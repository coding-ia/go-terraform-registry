package auth

import (
	"context"
	"github.com/google/go-github/v69/github"
	"log"
)

func GetGitHubUserName(ctx context.Context, client *github.Client, token string) (*string, error) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return nil, err
	}

	return user.Login, nil
}
