package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DBStorage struct {
	DB     *sql.DB
	Logger *zap.Logger
}

var ErrURLExists = fmt.Errorf("url already exists")

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
		DB:     db,
		Logger: logger,
	}, nil
}

func (s *DBStorage) SaveBatch(ctx context.Context, batch map[string]string) (map[string]string, map[string]string, error) {
	newMap := make(map[string]string)
	conflictMap := make(map[string]string)

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Logger.Error("failed to start transaction", zap.Error(err))
		return nil, nil, err
	}

	stmtInsert, err := tx.PrepareContext(ctx, `
        INSERT INTO urls (short_url, original_url)
        VALUES ($1, $2)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING short_url;
    `)
	if err != nil {
		s.Logger.Error("failed to prepare insert statement", zap.Error(err))
		tx.Rollback()
		return nil, nil, err
	}

	stmtSelectByOriginal, err := tx.PrepareContext(ctx, `
        SELECT short_url FROM urls WHERE original_url = $1;
    `)
	if err != nil {
		s.Logger.Error("failed to prepare select-original statement", zap.Error(err))
		tx.Rollback()
		return nil, nil, err
	}

	stmtSelectByShort, err := tx.PrepareContext(ctx, `
        SELECT original_url FROM urls WHERE short_url = $1;
    `)
	if err != nil {
		s.Logger.Error("failed to prepare select-short statement", zap.Error(err))
		tx.Rollback()
		return nil, nil, err
	}

	for shortID, origURL := range batch {
		var returnedID string
		err = stmtInsert.QueryRowContext(ctx, shortID, origURL).Scan(&returnedID)

		switch {
		case err == nil:
			newMap[origURL] = returnedID
			s.Logger.Debug("inserted new url", zap.String("short", returnedID), zap.String("original", origURL))

		case errors.Is(err, sql.ErrNoRows):
			var existing string
			if err := stmtSelectByOriginal.QueryRowContext(ctx, origURL).Scan(&existing); err == nil {
				conflictMap[origURL] = existing
				s.Logger.Debug("url already exists", zap.String("short", existing), zap.String("original", origURL))
			} else {
				s.Logger.Error("failed to resolve conflict by original url", zap.Error(err))
				tx.Rollback()
				return nil, nil, err
			}

		default:
			var existingOrig string
			if err := stmtSelectByShort.QueryRowContext(ctx, shortID).Scan(&existingOrig); err == nil {
				s.Logger.Error("shortID conflict", zap.String("shortID", shortID))
				tx.Rollback()
				return nil, nil, fmt.Errorf("shortID conflict: %s already taken", shortID)
			}

			s.Logger.Error("insert failed", zap.String("shortID", shortID), zap.String("original", origURL), zap.Error(err))
			tx.Rollback()
			return nil, nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		s.Logger.Error("failed to commit transaction", zap.Error(err))
		return nil, nil, err
	}

	s.Logger.Info("batch saved", zap.Int("new", len(newMap)), zap.Int("conflicts", len(conflictMap)))

	return newMap, conflictMap, nil
}

func (s *DBStorage) Save(ctx context.Context, id, url string) (string, bool, error) {
	query := `
       INSERT INTO urls (short_url, original_url)
        VALUES ($1, $2)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING short_url;
	`

	var savedID string
	err := s.DB.QueryRowContext(ctx, query, id, url).Scan(&savedID)

	switch {
	case err == nil:
		s.Logger.Info("Saved record", zap.String("short", savedID), zap.String("url", url))
		return savedID, true, nil

	case errors.Is(err, sql.ErrNoRows):
		var existingID string
		sel := `SELECT short_url FROM urls WHERE original_url = $1`
		if err := s.DB.QueryRowContext(ctx, sel, url).Scan(&existingID); err != nil {
			s.Logger.Error("conflict but cannot fetch existing short_url", zap.Error(err))
			return "", false, err
		}
		s.Logger.Info("URL already exists", zap.String("short", existingID), zap.String("url", url))
		return existingID, false, ErrURLExists

	default:
		s.Logger.Error("DB save failed", zap.Error(err))
		return "", false, err
	}
}

func (s *DBStorage) Get(id string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	var original string
	err := s.DB.QueryRowContext(ctx, query, id).Scan(&original)
	if err != nil {
		s.Logger.Error("Failed to get record from DB", zap.Error(err))
		return "", false
	}
	return original, true
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *DBStorage) Close() error {
	return s.DB.Close()
}
