package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type testObserver struct {
	mu     sync.Mutex
	events []Event
}

func (t *testObserver) Notify(_ context.Context, e Event) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)
	return nil
}

func TestService_Notify(t *testing.T) {
	log := zap.NewNop()

	obs1 := &testObserver{}
	obs2 := &testObserver{}

	svc := NewService(log, obs1, obs2)

	event := Event{
		TS:     time.Now().Unix(),
		Action: "shorten",
		UserID: "u1",
		URL:    "https://example.com",
	}

	svc.Notify(context.Background(), event)
	time.Sleep(20 * time.Millisecond)

	assert.Len(t, obs1.events, 1)
	assert.Len(t, obs2.events, 1)
}
