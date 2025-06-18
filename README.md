# GoWork

A Simple Go Worker Pool(runs single job multiple time).

gowork is a small Go library for running a single task across multiple goroutines (workers) at the same time.
You define one job, and gowork runs it for you on as many workers as you need, collecting any errors that happen along the way.

Features

- Run Concurrently: Run the same function many times in parallel.
- Error Collection: Collect all errors from every worker.
- Panic Recovery: If a worker panics, records the error and stack trace, and keeps the other workers running.
- Graceful Shutdown: Uses context to shut down all workers when you need them to stop.

Note: You will need to handle the contex passed by the `gowork` lib gracefully in your `task func`.

---

## Examples

1. Define Task
```go
import (
    "context"
    "fmt"
    "time"
    "github.com/prakashpandey/gowork"
)

// This is the job we want to run multiple times.
myTask := gowork.Task{
    Fn: func(ctx context.Context) error {
        fmt.Println("Worker is doing its job...")
        time.Sleep(10 * time.Second)
        return nil // or return error
    },
}
```

2. Create and Run the Pool

Next, create a SingleTaskPool, tell it how many workers you want, and start it.

```go
func main() {
    // Create a pool that will run the task on 5 workers at the same time.
    pool := gowork.NewPool(context.Background(), 5, myTask)

    // Start all the workers(Not a blocking call). 
    pool.Go()
    // Wait for all workers to finish their job(This is a blocking call).
    pool.Wait()
}
```

3. Error Handling

The pool collects every error from every worker.

```go
import (
    "context"
    "errors"
    "fmt"
    "github.com/prakashpandey/gowork"
)

func main() {
    failingTask := gowork.Task{
        Fn: func(ctx context.Context) error {
            return errors.New("something went wrong!")
        },
    }

    pool := gowork.NewPool(context.Background(), 3, failingTask)
    pool.Go()
    pool.Wait()

    // Check for errors after the pool is finished.
    errs := pool.Errors()
    if len(errs) > 0 {
        fmt.Printf("The pool finished with %d errors.\n", len(errs))
        for _, err := range errs {
            fmt.Printf("%s\n", err.Error())
        }
    } else {
        fmt.Println("The pool finished successfully with no errors.")
    }
}
```

---