package handler_test

import (
	"context"
	"fmt"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/audit"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"go.uber.org/zap"
)

func Example() {
	// Создаем in-memory хранилище
	store := storage.NewInMemoryStorage()

	// Создаем сервис аудита с nop-логером
	auditSvc := audit.NewService(zap.NewNop())

	baseURL := "http://localhost:8080"

	ctx := context.Background()

	// Сохраняем несколько URL (результат игнорируем через _)
	_, _ = store.Save(ctx, "user1", "abc123", "https://example.com")
	_, _ = store.Save(ctx, "user1", "def456", "https://golang.org")

	// Используем auditSvc чтобы уведомить о действиях
	auditSvc.Notify(ctx, audit.Event{TS: 1, Action: "shorten", UserID: "user1", URL: "https://example.com"})
	auditSvc.Notify(ctx, audit.Event{TS: 2, Action: "shorten", UserID: "user1", URL: "https://golang.org"})

	// Сохраняем batch URL
	batch := []storage.BatchItem{
		{ShortID: "ghi789", OriginalURL: "https://ya.ru"},
		{ShortID: "jkl012", OriginalURL: "https://google.com"},
	}
	store.SaveBatch(ctx, "user1", batch)

	// Получаем все URL пользователя
	urls, _ := store.GetUserURLs(ctx, "user1")
	for _, u := range urls {
		fmt.Println(baseURL+"/"+u.ShortID, u.OriginalURL)
	}

	// Output:
	// http://localhost:8080/abc123 https://example.com
	// http://localhost:8080/def456 https://golang.org
	// http://localhost:8080/ghi789 https://ya.ru
	// http://localhost:8080/jkl012 https://google.com
}
