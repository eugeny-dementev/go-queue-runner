# queue-runner

Minimal async queues for Go with functional actions, branching helpers, context mutation, and optional locking.

## Install

```sh
go get github.com/eugeny-dementev/go-queue-runner
```

## Quick start

```go
package main

import (
	"fmt"

	queuerunner "github.com/eugeny-dementev/go-queue-runner"
)

func main() {
	runner := queuerunner.NewQueueRunner(queuerunner.RunnerOpts{})

	actions := []queuerunner.Action{
		func(ctx *queuerunner.Context) error {
			ctx.Extend(map[string]any{"value": 1})
			return nil
		},
		func(ctx *queuerunner.Context) error {
			value := ctx.Data["value"].(int)
			if value >= 1 {
				ctx.Abort()
			}
			return nil
		},
	}

	runner.Add(actions, map[string]any{}, "")

	fmt.Println("queue started")
}
```

## Core concepts

### Action

Actions are plain functions:

```go
type Action func(*Context) error
```

### Context

Each action receives a mutable context:

```go
type Context struct {
	Data   map[string]any
	Logger Logger
	// Push, Extend, Name, Abort helpers
}
```

- `Push` inserts actions at the front of the remaining queue.
- `Extend` merges new fields into the context.
- `Name` returns the queue name.
- `Abort` clears the remaining queue.

### Error handling

By default, a returned error logs to `ctx.Logger` (or `ctx.Data["logger"]`) and aborts the queue.
You can override per action using `WithErrorHandler`:

```go
action := queuerunner.WithErrorHandler(func(_ *queuerunner.Context) error {
	return fmt.Errorf("boom")
}, func(_ error, ctx *queuerunner.Context) {
	ctx.Extend(map[string]any{"recovered": true})
})
```

To delay a single action, wrap it with `WithDelay`:

```go
action := queuerunner.WithDelay(func(_ *queuerunner.Context) error {
	return nil
}, 500*time.Millisecond)
```

### Locking

Use `WithLock` to serialize actions across queues by scope:

```go
action := queuerunner.WithLock("browser", func(ctx *queuerunner.Context) error {
	// protected by lock scope "browser"
	return nil
})
```

## Utilities

```go
// fixed delay
queuerunner.Util.Delay(500 * time.Millisecond)

// branching
queuerunner.Util.If(func(ctx *queuerunner.Context) (bool, error) {
	flag, _ := ctx.Data["flag"].(bool)
	return flag, nil
}, queuerunner.Branches{
	Then: []queuerunner.Action{someAction},
	Else: []queuerunner.Action{otherAction},
})

// conditional actions
queuerunner.Util.Valid(func(ctx *queuerunner.Context) (bool, error) {
	count, _ := ctx.Data["count"].(int)
	return count > 0, nil
}, []queuerunner.Action{someAction})

// immediate abort action
queuerunner.Util.Abort
```

## QueueRunner vs Queue

- `QueueRunner` manages multiple queues and shared locking.
- `Queue` runs a single queue directly.

```go
queue := queuerunner.NewQueue(queuerunner.QueueOpts{
	Name:    "my-queue",
	Actions: []queuerunner.Action{someAction},
})
queue.Run(map[string]any{"initial": true})
```

## Logging

`Queue` accepts a logger for queue-level logs.
The default error handler uses the context logger if present.

```go
logger := &MyLogger{}
runner := queuerunner.NewQueueRunner(queuerunner.RunnerOpts{Logger: logger})
runner.Add([]queuerunner.Action{someAction}, map[string]any{"logger": logger}, "")
```
