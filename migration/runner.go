package migration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(dbpool *pgxpool.Pool) error {
	ctx := context.Background()

	if err := createMigrationsTable(ctx, dbpool); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	files, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	sort.Strings(files)

	for _, file := range files {
		if err := applyMigration(ctx, dbpool, file); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}
	}

	log.Println("All database migrations completed successfully")
	return nil
}



func createMigrationsTable(ctx context.Context, dbpool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := dbpool.Exec(ctx, query)
	return err
}

func getMigrationFiles() ([]string, error) {
	dir := "migrations"
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(dir, file.Name()))
		}
	}
	return sqlFiles, nil
}

func applyMigration(ctx context.Context, dbpool *pgxpool.Pool, filePath string) error {
	version := strings.TrimSuffix(filepath.Base(filePath), ".sql")

	var exists bool
	err := dbpool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Migration %s already applied, skipping", version)
		return nil
	}

	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	sql := string(sqlBytes)

	_, err = dbpool.Exec(ctx, sql)
	if err != nil {
		return err
	}

	_, err = dbpool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		return err
	}

	log.Printf("Applied migration: %s", version)
	return nil
}
