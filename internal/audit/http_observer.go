package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPObserver отправляет события аудита на внешний HTTP-эндпоинт.
// Используется для интеграции с внешними системами логирования или мониторинга.
type HTTPObserver struct {
	url    string       // URL, на который отправляются события
	client *http.Client // HTTP-клиент для отправки запросов
	logger *zap.Logger  // Logger для ошибок и предупреждений
}

// NewHTTPObserver создаёт нового HTTPObserver с указанным URL и логгером.
// Если logger равен nil, используется zap.NewNop() — логгер-заглушка.
func NewHTTPObserver(url string, logger *zap.Logger) *HTTPObserver {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &HTTPObserver{
		url:    url,
		logger: logger,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Notify отправляет событие e на сконфигурированный URL.
// В случае ошибки при маршалинге или отправке запроса логирует ошибку через zap.Logger.
// Возвращает ошибку, если событие не удалось отправить.
func (o *HTTPObserver) Notify(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		o.logger.Error(
			"failed to marshal audit event",
			zap.Error(err),
			zap.String("action", event.Action),
			zap.String("user_id", event.UserID),
		)
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.url,
		bytes.NewReader(data),
	)
	if err != nil {
		o.logger.Error("failed to create audit http request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		o.logger.Warn(
			"failed to send audit event",
			zap.Error(err),
			zap.String("url", o.url),
		)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		o.logger.Warn(
			"audit service returned non-2xx status",
			zap.Int("status", resp.StatusCode),
			zap.String("url", o.url),
		)
	}

	return nil
}
