// Package storage содержит интерфейсы и структуры для работы с хранилищем URL.
// Определяет единый интерфейс Storage, который поддерживает различные реализации:
// в памяти, через файлы или через БД.
package storage

import "context"

// URLRecord представляет одну запись URL в хранилище.
// Используется как единая структура для всех типов хранилищ.
type URLRecord struct {
	// ShortID — короткий идентификатор URL.
	ShortID string
	// OriginalURL — оригинальный URL.
	OriginalURL string
	// UserID — идентификатор пользователя, которому принадлежит URL.
	UserID string
	// Deleted — флаг, помечающий URL как удалённый.
	Deleted bool
}

// BatchItem используется для пакетного сохранения URL.
type BatchItem struct {
	// ShortID — короткий идентификатор URL.
	ShortID string
	// OriginalURL — оригинальный URL.
	OriginalURL string
}

// Storage описывает интерфейс хранилища URL.
// Поддерживает как единичное, так и пакетное сохранение,
// получение URL по короткому идентификатору, список URL пользователя,
// а также пометку URL как удалённых.
type Storage interface {
	// Save сохраняет один URL для пользователя.
	// Возвращает короткий идентификатор и ошибку.
	Save(ctx context.Context, userID, id, url string) (string, error)

	// Get возвращает запись URL по короткому идентификатору.
	// Возвращает nil и false, если запись не найдена.
	Get(id string) (*URLRecord, bool)

	// SaveBatch сохраняет несколько URL одним батчем.
	// Возвращает новые URL и уже существующие.
	SaveBatch(ctx context.Context, userID string, batch []BatchItem) (map[string]string, map[string]string, error)

	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error

	// GetUserURLs возвращает список URL для указанного пользователя.
	GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error)

	// MarkDeleted помечает список URL как удалённые для указанного пользователя.
	MarkDeleted(userID string, shorts []string) error
}
