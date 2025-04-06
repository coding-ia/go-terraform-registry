package cli

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

type PostgresOptions struct {
	MigrationPath string
	DatabaseURL   string
}

var postgresOptions = &PostgresOptions{}

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Manage backend schema",
}

var postgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Manage postgres schema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		postgresMigrate(cmd.Context(), args)
	},
}

func init() {
	rootCmd.AddCommand(backendCmd)
	backendCmd.AddCommand(postgresCmd)

	postgresCmd.Flags().StringVar(&postgresOptions.MigrationPath, "migration-path", "", "Files for handling migration")
	postgresCmd.Flags().StringVar(&postgresOptions.DatabaseURL, "database", "", "Data connection string")
}

func postgresMigrate(_ context.Context, args []string) {
	action := args[0]
	fmt.Printf("Running migration: %s\n", action)

	filePath := fmt.Sprintf("file://%s", postgresOptions.MigrationPath)
	m, err := migrate.New(filePath, postgresOptions.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	if strings.EqualFold(action, "up") {
		err = m.Up()
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

		fmt.Println("Migration succeeded")
		return
	}

	if strings.EqualFold(action, "down") {
		err = m.Down()
		if err != nil {
			log.Fatalf("Downgrade failed: %v", err)
		}

		fmt.Println("Downgrade succeeded")
		return
	}
}
