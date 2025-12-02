package storage_test

import (
	"context"
	"database/sql"
	_ "errors"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDBStorage_Save(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &storage.DBStorage{
		DB:     db,
		Logger: logger,
	}

	ctx := context.Background()
	userID := "user123"

	t.Run("insert new URL", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO urls .* RETURNING short_url").
			WithArgs("short1", "https://example.com", userID).
			WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("short1"))

		shortID, err := s.Save(ctx, userID, "short1", "https://example.com")
		assert.NoError(t, err)
		assert.Equal(t, "short1", shortID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("URL already exists", func(t *testing.T) {

		mock.ExpectQuery("INSERT INTO urls .* RETURNING short_url").
			WithArgs("short2", "https://exists.com", userID).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url = \\$1").
			WithArgs("https://exists.com").
			WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("existing1"))

		shortID, err := s.Save(ctx, userID, "short2", "https://exists.com")
		assert.ErrorIs(t, err, storage.ErrURLExists)
		assert.Equal(t, "existing1", shortID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDBStorage_SaveBatch(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &storage.DBStorage{
		DB:     db,
		Logger: logger,
	}

	ctx := context.Background()
	userID := "user123"

	// Начало транзакции
	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO urls .* RETURNING short_url")
	mock.ExpectPrepare("SELECT short_url FROM urls WHERE original_url = \\$1")

	// успешная вставка
	mock.ExpectQuery("INSERT INTO urls .* RETURNING short_url").
		WithArgs("shortA", "https://a.com", userID).
		WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("shortA"))

	mock.ExpectCommit()

	batch := []storage.BatchItem{
		{ShortID: "shortA", OriginalURL: "https://a.com"},
	}

	newMap, conflictMap, err := s.SaveBatch(ctx, userID, batch)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"https://a.com": "shortA"}, newMap)
	assert.Empty(t, conflictMap)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_Get(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &storage.DBStorage{DB: db, Logger: logger}

	t.Run("existing ID", func(t *testing.T) {
		mock.ExpectQuery("SELECT original_url, user_id, is_deleted FROM urls WHERE short_url = \\$1").
			WithArgs("short1").
			WillReturnRows(sqlmock.NewRows([]string{"original_url", "user_id", "is_deleted"}).
				AddRow("https://example.com", "user123", false))

		rec, ok := s.Get("short1")
		assert.True(t, ok)
		assert.Equal(t, "https://example.com", rec.OriginalURL)
		assert.Equal(t, "user123", rec.UserID)
		assert.False(t, rec.Deleted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("non-existent ID", func(t *testing.T) {
		mock.ExpectQuery("SELECT original_url, user_id, is_deleted FROM urls WHERE short_url = \\$1").
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		rec, ok := s.Get("unknown")
		assert.False(t, ok)
		assert.Nil(t, rec)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDBStorage_GetUserURLs(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &storage.DBStorage{DB: db, Logger: logger}
	ctx := context.Background()
	userID := "user123"

	mock.ExpectQuery("SELECT short_url, original_url FROM urls WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"short_url", "original_url"}).
			AddRow("shortA", "https://a.com").
			AddRow("shortB", "https://b.com"))

	urls, err := s.GetUserURLs(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, urls, 2)
	assert.Equal(t, "shortA", urls[0].ShortID)
	assert.Equal(t, "https://a.com", urls[0].OriginalURL)
	assert.Equal(t, "shortB", urls[1].ShortID)
	assert.Equal(t, "https://b.com", urls[1].OriginalURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_MarkDeleted(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &storage.DBStorage{DB: db, Logger: logger}

	userID := "user123"
	shorts := []string{"shortA", "shortB"}

	mock.ExpectExec("UPDATE urls SET is_deleted = TRUE WHERE user_id = \\$1 AND short_url = ANY\\(\\$2\\)").
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(2, 2))

	err = s.MarkDeleted(userID, shorts)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
