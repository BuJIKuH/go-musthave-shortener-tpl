// Package service содержит общий фасад для HTTP и gRPC.
package service

import (
	"context"
	"errors"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service/shortener"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
)

var ErrInvalidURL = errors.New("invalid URL")

// ShortenerServiceFacade объединяет бизнес-логику сокращения URL.
type ShortenerServiceFacade struct {
	store storage.Storage
}

// NewShortenerFacade создает фасад для общего использования
func NewShortenerFacade(store storage.Storage) *ShortenerServiceFacade {
	return &ShortenerServiceFacade{store: store}
}

// ShortenURL сокращает URL и возвращает короткий идентификатор
func (f *ShortenerServiceFacade) ShortenURL(ctx context.Context, userID, url string) (string, error) {
	if url == "" {
		return "", ErrInvalidURL
	}

	id, err := shortener.GenerateID()
	if err != nil {
		return "", err
	}

	return f.store.Save(ctx, userID, id, url)
}

// ExpandURL возвращает оригинальный URL по короткому идентификатору
func (f *ShortenerServiceFacade) ExpandURL(ctx context.Context, id string) (string, error) {
	rec, ok := f.store.Get(id)
	if !ok || rec.Deleted {
		return "", errors.New("URL not found")
	}
	return rec.OriginalURL, nil
}

// ListUserURLs возвращает все URL пользователя
func (f *ShortenerServiceFacade) ListUserURLs(ctx context.Context, userID string) ([]storage.BatchItem, error) {
	return f.store.GetUserURLs(ctx, userID)
}
