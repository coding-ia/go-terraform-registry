package cli

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"time"
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

	migrations := os.Getenv("MIGRATIONS")
	dbUrl := os.Getenv("DATABASE_URL")
	postgresCmd.Flags().StringVar(&postgresOptions.MigrationPath, "migration-path", migrations, "Files for handling migration")
	postgresCmd.Flags().StringVar(&postgresOptions.DatabaseURL, "database", dbUrl, "Data connection string")

	if dbUrl != "" {
		_ = postgresCmd.MarkFlagRequired("database")
	}

	if migrations != "" {
		_ = postgresCmd.MarkFlagRequired("migrations")
	}
}

func postgresMigrate(ctx context.Context, args []string) {
	action := args[0]
	
	log.Printf("Running migration: %s", action)
	log.Printf("Migration path: %s", postgresOptions.MigrationPath)

	waitForPostgres(ctx, postgresOptions.DatabaseURL, 20)

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

		log.Println("Migration succeeded")
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

func waitForPostgres(ctx context.Context, connectionString string, maxRetries int) {
	for i := 0; i < maxRetries; i++ {
		available := testConnection(ctx, connectionString)
		if available {
			return
		}
		log.Printf("Waiting for PostgreSQL... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(3 * time.Second)
	}

	log.Println("PostgreSQL not ready after max retries.")
	os.Exit(1)
}

func testConnection(ctx context.Context, connectionString string) bool {
	conn, err := pgx.Connect(ctx, connectionString)
	defer func() {
		if conn != nil {
			_ = conn.Close(ctx)
		}
	}()
	if err == nil {
		return true
	}
	return false
}
