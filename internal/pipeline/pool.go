package pipeline

import (
	"context"
	"sync"
)

// WorkerPool manages bounded concurrency for network tasks.
type WorkerPool struct {
	sem chan struct{}
	wg  sync.WaitGroup
	ctx context.Context
}

// NewWorkerPool creates a worker pool bounded to maxConcurrency.
func NewWorkerPool(ctx context.Context, maxConcurrency int) *WorkerPool {
	if maxConcurrency <= 0 {
		maxConcurrency = 20
	}
	return &WorkerPool{
		sem: make(chan struct{}, maxConcurrency),
		ctx: ctx,
	}
}

// Submit enqueues a task into the bounded pool. If the context is canceled, task is dropped.
func (wp *WorkerPool) Submit(task func()) bool {
	select {
	case <-wp.ctx.Done():
		return false
	case wp.sem <- struct{}{}:
	}

	wp.wg.Add(1)
	go func() {
		defer func() {
			<-wp.sem
			wp.wg.Done()
		}()

		// Recover from unexpected panics in individual workers to prevent whole process crash
		defer func() {
			if r := recover(); r != nil {
				// Worker panic isolated gracefully
			}
		}()

		task()
	}()

	return true
}

// Wait blocks until all active worker tasks have finished.
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
