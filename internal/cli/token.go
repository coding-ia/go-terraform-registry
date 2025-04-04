package cli

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"go-terraform-registry/internal/auth"
	"os"
)

type TokenOptions struct {
	Key  string
	User string
}

var tokenOptions = &TokenOptions{}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "API tokens",
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate API tokens",
	Run: func(cmd *cobra.Command, args []string) {
		generateToken(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(generateCmd)

	tokenKey := os.Getenv("")
	generateCmd.Flags().StringVar(&tokenOptions.Key, "key", tokenKey, "Token encryption key")
	generateCmd.Flags().StringVar(&tokenOptions.User, "user", "", "User name")

	_ = generateCmd.MarkFlagRequired("user")
	if tokenKey == "" {
		_ = generateCmd.MarkFlagRequired("key")
	}
}

func generateToken(_ context.Context) {
	token, err := auth.CreateJWTToken(tokenOptions.User, []byte(tokenOptions.Key))
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		return
	}
	if token != nil {
		fmt.Printf("Generated token: %s\n", *token)
	}
}
