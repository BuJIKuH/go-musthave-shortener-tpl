package storage

import "context"

// URLRecord — единая запись для всех хранилищ
type URLRecord struct {
	ShortID     string
	OriginalURL string
	UserID      string
	Deleted     bool
}

type BatchItem struct {
	ShortID     string
	OriginalURL string
}

// Storage — обновлённый интерфейс
type Storage interface {
	Save(ctx context.Context, userID, id, url string) (string, error)
	Get(id string) (*URLRecord, bool)
	SaveBatch(ctx context.Context, userID string, batch []BatchItem) (map[string]string, map[string]string, error)
	Ping(ctx context.Context) error
	GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error)
	MarkDeleted(userID string, shorts []string) error
}
