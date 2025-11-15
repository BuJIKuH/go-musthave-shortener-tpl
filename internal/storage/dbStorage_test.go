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

	t.Run("insert new URL", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO urls .* RETURNING short_url").
			WithArgs("short1", "https://example.com").
			WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("short1"))

		shortID, conflict, err := s.Save(ctx, "short1", "https://example.com")
		assert.NoError(t, err)
		assert.True(t, conflict) // Save возвращает true для новой записи
		assert.Equal(t, "short1", shortID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("URL already exists", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO urls .* RETURNING short_url").
			WithArgs("short2", "https://exists.com").
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url = \\$1").
			WithArgs("https://exists.com").
			WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("existing1"))

		shortID, conflict, err := s.Save(ctx, "short2", "https://exists.com")
		assert.ErrorIs(t, err, storage.ErrURLExists)
		assert.False(t, conflict)
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

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO urls .* RETURNING short_url")
	mock.ExpectPrepare("SELECT short_url FROM urls WHERE original_url = \\$1")
	mock.ExpectPrepare("SELECT original_url FROM urls WHERE short_url = \\$1")

	// Мок QueryRow для вставки
	mock.ExpectQuery("INSERT INTO urls .* RETURNING short_url").
		WithArgs("shortA", "https://a.com").
		WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("shortA"))

	mock.ExpectCommit()

	batch := map[string]string{"shortA": "https://a.com"}
	newMap, conflictMap, err := s.SaveBatch(ctx, batch)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"https://a.com": "shortA"}, newMap)
	assert.Empty(t, conflictMap)
	assert.NoError(t, mock.ExpectationsWereMet())

}
