package audit

import "context"

type Service struct {
	observers []Observer
	events    chan Event
}

func NewService(observers ...Observer) *Service {
	s := &Service{
		observers: observers,
		events:    make(chan Event, 1000),
	}

	go func() {
		for event := range s.events {
			for _, obs := range s.observers {
				_ = obs.Notify(context.Background(), event)
			}
		}
	}()

	return s
}

func (s *Service) Notify(ctx context.Context, e Event) {
	select {
	case s.events <- e:
	default:
		// очередь полна, можно логировать потерю
	}
}
