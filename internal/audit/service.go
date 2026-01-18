package audit

import "context"

type Service struct {
	observers []Observer
}

func NewService(observers ...Observer) *Service {
	return &Service{observers: observers}
}

func (s *Service) Notify(ctx context.Context, event Event) {
	for _, obs := range s.observers {

		_ = obs.Notify(ctx, event)
	}
}
