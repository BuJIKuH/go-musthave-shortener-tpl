package audit

import (
	"context"

	"go.uber.org/zap"
)

type Service struct {
	observers []Observer
	events    chan Event
	logger    *zap.Logger
}

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
