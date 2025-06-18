package gowork

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

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
