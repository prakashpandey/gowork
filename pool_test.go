package gowork

import (
	"context"
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
