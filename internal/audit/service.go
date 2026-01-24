package audit

import (
	"context"

	"go.uber.org/zap"
)

// Service обеспечивает асинхронное уведомление наблюдателей об аудиторских событиях.
// События помещаются в буферизированный канал и обрабатываются в отдельной горутине.
// Если канал переполнен, событие теряется с записью предупреждения в логгер.
type Service struct {
	observers []Observer  // список наблюдателей
	events    chan Event  // буферизированный канал для событий
	logger    *zap.Logger // логгер для ошибок и предупреждений
}

// NewService создаёт новый сервис аудита с указанным логгером и списком наблюдателей.
// Логгер не может быть nil, иначе паника при попытке логирования.
// Обработчик событий запускается в отдельной горутине.
func NewService(logger *zap.Logger, observers ...Observer) *Service {
	s := &Service{
		observers: observers,
		events:    make(chan Event, 1000),
		logger:    logger,
	}

	go func() {
		for event := range s.events {
			for _, obs := range s.observers {
				if err := obs.Notify(context.Background(), event); err != nil {
					s.logger.Error(
						"audit observer notify failed",
						zap.Error(err),
						zap.String("action", event.Action),
						zap.String("user_id", event.UserID),
						zap.String("url", event.URL),
						zap.Int64("ts", event.TS),
					)
				}
			}
		}
	}()

	return s
}

// Notify помещает событие e в очередь для асинхронной отправки наблюдателям.
// Если очередь переполнена, событие теряется, а предупреждение логируется.
func (s *Service) Notify(ctx context.Context, e Event) {
	select {
	case s.events <- e:
	default:
		s.logger.Warn(
			"audit event dropped: queue is full",
			zap.String("action", e.Action),
			zap.String("user_id", e.UserID),
			zap.String("url", e.URL),
			zap.Int64("ts", e.TS),
			zap.Int("queue_size", len(s.events)),
		)
	}
}
