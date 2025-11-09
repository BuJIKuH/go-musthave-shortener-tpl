package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DBStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDBStorage(dsn string, logger *zap.Logger) (*DBStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("сannot open DB: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot connect to DB: %w", err)
	}

	logger.Info("Connected to PostgreSQL successfully")

	return &DBStorage{
		db:     db,
		logger: logger,
	}, nil
}

func (s *DBStorage) Save(id, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	query := `INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`
	_, err := s.db.ExecContext(ctx, query, id, url)
	if err != nil {
		s.logger.Error("Failed to save record to DB", zap.Error(err))
		return
	}
	s.logger.Info("Saved record", zap.String("short", id), zap.String("url", url))
}

func (s *DBStorage) Get(id string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	var original string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&original)
	if err != nil {
		s.logger.Error("Failed to get record from DB", zap.Error(err))
		return "", false
	}
	return original, true
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
