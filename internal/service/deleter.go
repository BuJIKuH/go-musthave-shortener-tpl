package service

import (
	"sync"
	"sync/atomic"
	"time"
)

type DeleteTask struct {
	UserID string
	IDs    []string
}

type Deleter struct {
	markFunc     func(userID string, shorts []string) error
	queues       []chan DeleteTask
	fanIn        chan DeleteTask
	maxBatchSize int
	batchTimeout time.Duration

	done    chan struct{}
	wg      sync.WaitGroup
	counter uint32
}

func NewDeleter(markFunc func(userID string, shorts []string) error) *Deleter {
	const workers = 3

	d := &Deleter{
		markFunc:     markFunc,
		maxBatchSize: 100,
		batchTimeout: 200 * time.Millisecond,
		done:         make(chan struct{}),
		fanIn:        make(chan DeleteTask, 2048),
		queues:       make([]chan DeleteTask, workers),
	}

	for i := 0; i < workers; i++ {
		d.queues[i] = make(chan DeleteTask, 256)

		d.wg.Add(1)
		go func(ch <-chan DeleteTask) {
			defer d.wg.Done()

			for {
				select {
				case t, ok := <-ch:
					if !ok {
						return
					}
					select {
					case d.fanIn <- t:
					case <-d.done:
						return
					}

				case <-d.done:
					return
				}
			}
		}(d.queues[i])
	}

	d.wg.Add(1)
	go d.batchWorker()

	return d
}

func (d *Deleter) Enqueue(t DeleteTask) {
	if len(d.queues) == 0 {
		return
	}

	i := atomic.AddUint32(&d.counter, 1)
	idx := int(i % uint32(len(d.queues)))

	select {
	case d.queues[idx] <- t:
	default:
	}
}

func (d *Deleter) batchWorker() {
	defer d.wg.Done()

	buffer := make([]DeleteTask, 0, d.maxBatchSize*2)

	flush := func(tasks []DeleteTask) {
		if len(tasks) == 0 {
			return
		}
		group := make(map[string][]string)
		for _, t := range tasks {
			group[t.UserID] = append(group[t.UserID], t.IDs...)
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
		case <-d.done:
			flush(buffer)
			close(d.fanIn)
			return

		case t := <-d.fanIn:
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

func (d *Deleter) Close() {
	close(d.done)

	for _, q := range d.queues {
		close(q)
	}

	d.wg.Wait()
}
