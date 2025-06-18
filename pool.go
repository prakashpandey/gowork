package gowork

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

// Task defines the function to be executed by workers in the pool.
type Task struct {
	Fn func(ctx context.Context) error
}

// Error records an error from a specific worker, including
// the stack trace in case of a panic.
type Error struct {
	WorkerID int
	Err      error
	Stack    []byte
}

func (e *Error) Error() string {
	errStr := fmt.Sprintf("worker_id: %d, error: %v", e.WorkerID, e.Err)
	if len(e.Stack) > 0 {
		errStr = fmt.Sprintf("%s\nstacktrace:\n%s", errStr, e.Stack)
	}
	return errStr
}

// SingleTaskPool: A pool of goroutines executing a single task concurrently.
type SingleTaskPool struct {
	ctx        context.Context
	cancel     func()
	numWorkers int
	wg         sync.WaitGroup
	mu         sync.RWMutex
	task       Task
	errs       []Error
}

// NewPool creates a new worker pool.
// Takes a parent context, the num of concurrent workers and the task to run.
func NewPool(ctx context.Context, numWorkers int, task Task) *SingleTaskPool {
	childCtx, cancel := context.WithCancel(ctx)
	return &SingleTaskPool{
		ctx:        childCtx,
		cancel:     cancel,
		numWorkers: numWorkers,
		task:       task,
		// Initialize with 0 length but with a capacity to avoid re-allocations.
		errs: make([]Error, 0, numWorkers),
	}
}

// registerError adds(safely) an error from a worker to the pool error list.
func (pool *SingleTaskPool) registerError(err Error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.errs = append(pool.errs, err)
}

// Go starts all worker goroutines. Each worker will execute the task function.
func (pool *SingleTaskPool) Go() {
	for workerID := 0; workerID < pool.numWorkers; workerID++ {
		pool.wg.Add(1)
		// We are passing the workerID to avoid it being re-initialized on each for-loop.
		go func(workerID int) {
			defer pool.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					pool.registerError(Error{
						WorkerID: workerID,
						Err:      fmt.Errorf("panic recovered with error message: %v", r),
						Stack:    debug.Stack(),
					})
				}
			}()
			if err := pool.task.Fn(pool.ctx); err != nil {
				pool.registerError(Error{
					WorkerID: workerID,
					Err:      err,
				})
			}
		}(workerID)
	}
}

// Wait blocks until all worker goroutines have completed their execution.
func (p *SingleTaskPool) Wait() {
	p.wg.Wait()
	p.cancel()
}

// Errors returns a slice of all errors collected from the workers.
func (p *SingleTaskPool) Errors() []Error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// Return a copy to prevent race conditions on the slice
	// if the caller modifies it.
	errsCopy := make([]Error, len(p.errs))
	copy(errsCopy, p.errs)
	return errsCopy
}
