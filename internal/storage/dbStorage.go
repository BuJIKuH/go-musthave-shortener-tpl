// Package storage реализует хранение URL в PostgreSQL.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// DBStorage реализует хранение URL в PostgreSQL.
type DBStorage struct {
	DB     *sql.DB
	Logger *zap.Logger
}

// ErrURLExists возвращается, если сохраняемый URL уже существует.
var ErrURLExists = fmt.Errorf("url already exists")

// NewDBStorage создаёт новое подключение к базе данных PostgreSQL.
// Параметры:
//   - dsn: Data Source Name для подключения к БД.
//   - logger: zap.Logger для логирования операций.
//
// Возвращает:
//   - *DBStorage: готовый объект для работы с хранилищем.
//   - error: ошибка при открытии или пинге БД.
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

// SaveBatch сохраняет несколько URL в хранилище одним батчем.
// Параметры:
//   - ctx: context для контроля таймаута и отмены.
//   - userID: идентификатор пользователя.
//   - batch: список элементов BatchItem для сохранения.
//
// Возвращает:
//   - map[string]string: новые URL и их короткие идентификаторы.
//   - map[string]string: URL, которые уже существовали (conflict).
//   - error: ошибка выполнения операции.
func (s *DBStorage) SaveBatch(ctx context.Context, userID string, batch []BatchItem) (map[string]string, map[string]string, error) {
	newMap := make(map[string]string)
	conflictMap := make(map[string]string)

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	stmtInsert, err := tx.PrepareContext(ctx, `
        INSERT INTO urls (short_url, original_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING short_url;
    `)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	stmtSelect, err := tx.PrepareContext(ctx, `
        SELECT short_url FROM urls WHERE original_url = $1;
    `)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	for _, item := range batch {
		var existing string
		err := stmtInsert.QueryRowContext(ctx, item.ShortID, item.OriginalURL, userID).Scan(&existing)

		switch {
		case err == nil:
			newMap[item.OriginalURL] = existing

		case errors.Is(err, sql.ErrNoRows):
			err = stmtSelect.QueryRowContext(ctx, item.OriginalURL).Scan(&existing)
			if err != nil {
				tx.Rollback()
				return nil, nil, err
			}
			conflictMap[item.OriginalURL] = existing

		default:
			tx.Rollback()
			return nil, nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return newMap, conflictMap, nil
}

// Save сохраняет один URL в хранилище.
// Параметры:
//   - ctx: context запроса.
//   - userID: идентификатор пользователя.
//   - id: короткий идентификатор URL.
//   - url: оригинальный URL.
//
// Возвращает:
//   - string: короткий идентификатор, который был сохранён или уже существовал.
//   - error: ErrURLExists если URL уже существует, или другую ошибку.
func (s *DBStorage) Save(ctx context.Context, userID, id, url string) (string, error) {
	query := `
        INSERT INTO urls (short_url, original_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING short_url;
    `

	var savedID string
	err := s.DB.QueryRowContext(ctx, query, id, url, userID).Scan(&savedID)

	switch {
	case err == nil:
		return savedID, nil

	case errors.Is(err, sql.ErrNoRows):
		var existingID string
		sel := `SELECT short_url FROM urls WHERE original_url = $1`
		if err := s.DB.QueryRowContext(ctx, sel, url).Scan(&existingID); err != nil {
			return "", err
		}
		return existingID, ErrURLExists

	default:
		return "", err
	}
}

// Get возвращает запись URL по короткому идентификатору.
// Параметры:
//   - id: короткий идентификатор URL.
//
// Возвращает:
//   - *URLRecord: запись URL с полями ShortID, OriginalURL, UserID, Deleted.
//   - bool: true если запись найдена, false если не найдена.
func (s *DBStorage) Get(id string) (*URLRecord, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	query := `SELECT original_url, user_id, is_deleted FROM urls WHERE short_url = $1`
	var original, userID string
	var isDeleted bool
	err := s.DB.QueryRowContext(ctx, query, id).Scan(&original, &userID, &isDeleted)
	if err != nil {
		s.Logger.Debug("Get: not found or db error", zap.String("id", id), zap.Error(err))
		return nil, false
	}
	rec := &URLRecord{
		ShortID:     id,
		OriginalURL: original,
		UserID:      userID,
		Deleted:     isDeleted,
	}
	return rec, true
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *DBStorage) Close() error {
	return s.DB.Close()
}

// GetUserURLs возвращает список URL для указанного пользователя.
// Параметры:
//   - ctx: context запроса.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - []BatchItem: список коротких и оригинальных URL.
//   - error: ошибка запроса к базе.
func (s *DBStorage) GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT short_url, original_url FROM urls WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []BatchItem
	for rows.Next() {
		var it BatchItem
		if err := rows.Scan(&it.ShortID, &it.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// MarkDeleted помечает список URL как удалённые для указанного пользователя.
// Параметры:
//   - userID: идентификатор пользователя.
//   - shorts: список коротких идентификаторов для удаления.
//
// Возвращает:
//   - error: ошибка обновления записей в базе.
func (s *DBStorage) MarkDeleted(userID string, shorts []string) error {
	if len(shorts) == 0 {
		return nil
	}

	query := `
        UPDATE urls
        SET is_deleted = TRUE
        WHERE user_id = $1 AND short_url = ANY($2)
    `
	_, err := s.DB.Exec(query, userID, pq.Array(shorts))
	return err
}
