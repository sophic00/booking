package db

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// RunMigrations runs all pending up migrations.
func RunMigrations(databaseURL string) error {
	m, err := newMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("✅ Database migrations applied successfully")
	return nil
}

// RollbackMigration rolls back the last migration step.
func RollbackMigration(databaseURL string) error {
	m, err := newMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Println("✅ Rolled back 1 migration step successfully")
	return nil
}

// ForceVersion sets a migration version and clears the dirty flag.
func ForceVersion(databaseURL string, version int) error {
	m, err := newMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version %d: %w", version, err)
	}

	log.Printf("✅ Forced migration version to %d", version)
	return nil
}

// GetMigrationVersion returns the current version and dirty state.
func GetMigrationVersion(databaseURL string) (uint, bool, error) {
	m, err := newMigrator(databaseURL)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}

func newMigrator(databaseURL string) (*migrate.Migrate, error) {
	d, err := iofs.New(MigrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create iofs migration driver: %w", err)
	}

	driverURL := databaseURL
	if len(databaseURL) >= 11 && databaseURL[:11] == "postgres://" {
		driverURL = "pgx5://" + databaseURL[11:]
	} else if len(databaseURL) >= 13 && databaseURL[:13] == "postgresql://" {
		driverURL = "pgx5://" + databaseURL[13:]
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, driverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return m, nil
}
