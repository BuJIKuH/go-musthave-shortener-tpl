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
	// Параметры:
	//   - ctx: context запроса для контроля таймаута и отмены.
	//   - userID: идентификатор пользователя.
	//   - id: короткий идентификатор для URL.
	//   - url: оригинальный URL.
	// Возвращает:
	//   - string: короткий идентификатор сохранённого или уже существующего URL.
	//   - error: ErrURLExists если URL уже существует, либо другую ошибку.
	Save(ctx context.Context, userID, id, url string) (string, error)

	// Get возвращает запись URL по короткому идентификатору.
	// Параметры:
	//   - id: короткий идентификатор URL.
	// Возвращает:
	//   - *URLRecord: запись URL.
	//   - bool: true если запись найдена, false если не найдена.
	Get(id string) (*URLRecord, bool)

	// SaveBatch сохраняет несколько URL одним батчем.
	// Параметры:
	//   - ctx: context запроса.
	//   - userID: идентификатор пользователя.
	//   - batch: список элементов BatchItem для сохранения.
	// Возвращает:
	//   - map[string]string: новые URL и их короткие идентификаторы.
	//   - map[string]string: URL, которые уже существовали.
	//   - error: ошибка выполнения операции.
	SaveBatch(ctx context.Context, userID string, batch []BatchItem) (map[string]string, map[string]string, error)

	// Ping проверяет доступность хранилища.
	// Параметры:
	//   - ctx: context запроса.
	// Возвращает:
	//   - error: ошибка соединения, если есть.
	Ping(ctx context.Context) error

	// GetUserURLs возвращает список URL для указанного пользователя.
	// Параметры:
	//   - ctx: context запроса.
	//   - userID: идентификатор пользователя.
	// Возвращает:
	//   - []BatchItem: список коротких и оригинальных URL.
	//   - error: ошибка запроса.
	GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error)

	// MarkDeleted помечает список URL как удалённые для указанного пользователя.
	// Параметры:
	//   - userID: идентификатор пользователя.
	//   - shorts: список коротких идентификаторов для удаления.
	// Возвращает:
	//   - error: ошибка обновления записей в хранилище.
	MarkDeleted(userID string, shorts []string) error
}
