// Package storage содержит функции для работы с базой данных и миграциями.
package storage

import (
	"database/sql"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"go.uber.org/zap"
)

// RunMigrations выполняет миграции базы данных PostgreSQL.
// dns — строка подключения к базе данных.
// logger — zap.Logger для логирования ошибок и информации.
func RunMigrations(dns string, logger *zap.Logger) error {
	db, err := sql.Open("postgres", dns)
	if err != nil {
		logger.Error("cannot open DB", zap.Error(err))
		return err
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			logger.Error("failed to close DB connection", zap.Error(cerr))
		}
	}()

	path, err := filepath.Abs("internal/storage/migrations")
	if err != nil {
		logger.Error("failed to get absolute path", zap.Error(err))
		return err
	}

	migrationsURL := "file://" + path
	logger.Info("Running migrations", zap.String("path", migrationsURL))

	m, err := migrate.New(migrationsURL, dns)
	if err != nil {
		logger.Error("cannot create migration", zap.Error(err))
		return err
	}

	if upErr := m.Up(); upErr != nil && upErr != migrate.ErrNoChange {
		logger.Error("cannot run migration", zap.Error(upErr))
		return upErr
	}

	logger.Info("migrations successfully migrated")
	return nil
}
