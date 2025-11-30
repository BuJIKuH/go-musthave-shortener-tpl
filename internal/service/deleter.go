package service

import (
	"time"
)

type DeleteTask struct {
	UserID string
	IDs    []string
}

type Deleter struct {
	markFunc     func(userID string, shorts []string) error
	queue        chan DeleteTask
	maxBatchSize int
	batchTimeout time.Duration
}

func NewDeleter(markFunc func(userID string, shorts []string) error) *Deleter {
	d := &Deleter{
		markFunc:     markFunc,
		queue:        make(chan DeleteTask, 1024),
		maxBatchSize: 100,
		batchTimeout: 200 * time.Millisecond,
	}
	go d.worker()
	return d
}

func (d *Deleter) Enqueue(t DeleteTask) {
	select {
	case d.queue <- t:
	default:
	}
}

func (d *Deleter) worker() {
	buffer := make([]DeleteTask, 0, d.maxBatchSize*2)

	flush := func(tasks []DeleteTask) {
		if len(tasks) == 0 {
			return
		}
		group := make(map[string][]string)
		for _, t := range tasks {
			for _, id := range t.IDs {
				group[t.UserID] = append(group[t.UserID], id)
			}
		}
		for uid, ids := range group {
			_ = d.markFunc(uid, ids)
		}
	}

	timer := time.NewTimer(d.batchTimeout)
	defer timer.Stop()

	for {
		timer.Reset(d.batchTimeout)
		select {
		case t, ok := <-d.queue:
			if !ok {
				flush(buffer)
				return
			}
			buffer = append(buffer, t)
			if len(buffer) >= d.maxBatchSize {
				flush(buffer)
				buffer = buffer[:0]
			}
		case <-timer.C:
			if len(buffer) > 0 {
				flush(buffer)
				buffer = buffer[:0]
			}
		}
	}
}
