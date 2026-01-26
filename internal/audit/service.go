package audit

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Service обеспечивает асинхронное уведомление наблюдателей об аудиторских событиях.
// События помещаются в буферизированный канал и обрабатываются в отдельной горутине.
// Если канал переполнен, событие отбрасывается с записью предупреждения в лог.
type Service struct {
	observers []Observer  // список наблюдателей
	events    chan Event  // буферизированный канал событий
	logger    *zap.Logger // логгер для ошибок и предупреждений
}

// NewService создаёт новый сервис аудита с указанным логгером и наблюдателями.
// Для обработки событий запускается отдельная горутина.
// logger не должен быть nil.
func NewService(logger *zap.Logger, observers ...Observer) *Service {
	s := &Service{
		observers: observers,
		events:    make(chan Event, 1000),
		logger:    logger,
	}

	go s.run()

	return s
}

// run запускает основной цикл обработки событий.
func (s *Service) run() {
	for event := range s.events {
		for _, obs := range s.observers {
			// Защищаемся от зависаний observer'ов
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			if err := obs.Notify(ctx, event); err != nil {
				s.logger.Error(
					"audit observer notify failed",
					zap.Error(err),
					zap.String("action", event.Action),
					zap.String("user_id", event.UserID),
					zap.String("url", event.URL),
					zap.Int64("ts", event.TS),
				)
			}

			cancel()
		}
	}
}

// Notify помещает событие e в очередь для асинхронной отправки наблюдателям.
// Если очередь переполнена, событие отбрасывается, а предупреждение логируется.
//
// Контекст ctx используется только для управления жизненным циклом вызова Notify
// и не передаётся напрямую в observer'ы.
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
