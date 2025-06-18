package gowork

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSuccessfulExecution verifies that all workers run without error and the pool
// waits correctly.
func TestSuccessfulExecution(t *testing.T) {
	var counter int32
	task := Task{
		Fn: func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		},
	}
	pool := NewPool(context.Background(), 10, task)
	pool.Go()
	pool.Wait()

	assert.EqualValues(t, 10, counter)
	assert.EqualValues(t, 0, len(pool.Errors()))
}

func TestErrorCollections(t *testing.T) {
	numWorkers := 50
	task := Task{
		Fn: func(ctx context.Context) error {
			return errors.New("task failed")
		},
	}

	pool := NewPool(context.Background(), numWorkers, task)
	pool.Go()
	pool.Wait()

	errs := pool.Errors()
	assert.EqualValues(t, numWorkers, len(errs))

	workerIDs := make(map[int]bool)
	for _, e := range errs {
		// should not have duplicate worker id.
		assert.Equal(t, false, workerIDs[e.WorkerID])
		workerIDs[e.WorkerID] = true
	}
	assert.Equal(t, numWorkers, len(workerIDs))
}

func TestPanicRecovery(t *testing.T) {
	panicMsg := "worker panicked"
	task := Task{
		Fn: func(ctx context.Context) error {
			panic(panicMsg)
		},
	}
	pool := NewPool(context.Background(), 1, task)
	pool.Go()
	pool.Wait()

	assert.Contains(t, pool.Errors()[0].Error(), panicMsg)
}

func TestParentContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	teask := Task{
		Fn: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				return errors.New("worker didn't terminate on context cancellation")
			}
		},
	}

	pool := NewPool(ctx, 5, teask)
	pool.Go()
	time.Sleep(2 * time.Second)
	cancel()
	// Calling pool.Wait() after the cancel(): because pool.Wait() is a blocking call.
	pool.Wait()
	// We should not get any error as after calling the cancel() func, it returns 'nil'.
	assert.Equal(t, 0, len(pool.Errors()))
}

func TestZeroWorkers(t *testing.T) {
	task := Task{
		Fn: func(ctx context.Context) error {
			t.Error("task should not run with zero workers")
			return nil
		},
	}

	pool := NewPool(context.Background(), 0, task)
	pool.Go()
	pool.Wait()

	assert.Equal(t, 0, len(pool.Errors()))
}

func TestErrorMethodFormat(t *testing.T) {
	// Without a stack trace
	errNoStack := Error{
		WorkerID: 42,
		Err:      errors.New("simple error 1"),
		Stack:    nil,
	}
	expectedNoStack := "worker_id: 42, error: simple error 1"
	assert.Equal(t, expectedNoStack, errNoStack.Error())

	// With a stack trace
	errWithStack := Error{
		WorkerID: 101,
		Err:      errors.New("panic error"),
		Stack:    []byte("...some stack trace..."),
	}
	expectedWithStack := "worker_id: 101, error: panic error\nstacktrace:\n...some stack trace..."
	assert.Equal(t, expectedWithStack, errWithStack.Error())
}
