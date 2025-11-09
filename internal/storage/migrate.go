package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"go.uber.org/zap"
)

func RunMigrations(dns string, logger *zap.Logger) error {
	db, err := sql.Open("postgres", dns)
	if err != nil {
		logger.Fatal("cannot open DB", zap.Error(err))
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Info("cannot create postgres driver", zap.Error(err))
		return err
	}

	migrationsPath, _ := filepath.Abs("./internal/storage/migrations")
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath), "postgres", driver)
	if err != nil {
		logger.Info("cannot create migration", zap.Error(err))
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Info("cannot run migration", zap.Error(err))
		return err
	}
	logger.Info("migrations successfully migrated")

	return nil
}
