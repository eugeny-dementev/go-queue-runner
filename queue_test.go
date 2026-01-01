package queuerunner

import (
	"sync"
	"testing"
	"time"
)

func anyAction(execute func(ctx *Context) error) Action {
	return execute
}

func lockingAction(scope string, execute func(ctx *Context) error) Action {
	return WithLock(scope, execute)
}

func TestQueueOrderAndAbort(t *testing.T) {
	order := []string{}
	ended := false

	queue := NewQueue(QueueOpts{
		Actions: []Action{
			anyAction(func(_ *Context) error { order = append(order, "action"); return nil }),
			anyAction(func(_ *Context) error { order = append(order, "action"); return nil }),
			anyAction(func(_ *Context) error { order = append(order, "action"); return nil }),
			anyAction(func(ctx *Context) error { ctx.Abort(); return nil }),
			anyAction(func(_ *Context) error { order = append(order, "action"); return nil }),
		},
		Name: "TestQueue",
		End: func() {
			ended = true
		},
		LockingContext: NewLockManager(),
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{})

	if !ended {
		t.Fatal("queue did not end")
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 actions, got %v", order)
	}
}

func TestQueueLocking(t *testing.T) {
	lockingContext := NewLockManager()
	logger := &testLogger{}

	var mu sync.Mutex
	scopedCounter := 0
	errors := []error{}

	action := func() Action {
		return lockingAction("browser", func(_ *Context) error {
			mu.Lock()
			scopedCounter++
			current := scopedCounter
			mu.Unlock()

			if !lockingContext.IsLocked("browser") {
				mu.Lock()
				errors = append(errors, ErrInvalidScope)
				mu.Unlock()
			}
			if current != 1 {
				mu.Lock()
				errors = append(errors, ErrInvalidScope)
				mu.Unlock()
			}

			time.Sleep(5 * time.Millisecond)

			mu.Lock()
			scopedCounter--
			mu.Unlock()
			return nil
		})
	}

	queue1 := NewQueue(QueueOpts{
		Actions:        []Action{action()},
		Name:           "Q1",
		End:            func() {},
		LockingContext: lockingContext,
		Logger:         logger,
	})
	queue2 := NewQueue(QueueOpts{
		Actions:        []Action{action()},
		Name:           "Q2",
		End:            func() {},
		LockingContext: lockingContext,
		Logger:         logger,
	})
	queue3 := NewQueue(QueueOpts{
		Actions:        []Action{action()},
		Name:           "Q3",
		End:            func() {},
		LockingContext: lockingContext,
		Logger:         logger,
	})

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); queue1.Run(map[string]any{}) }()
	go func() { defer wg.Done(); queue2.Run(map[string]any{}) }()
	go func() { defer wg.Done(); queue3.Run(map[string]any{}) }()
	wg.Wait()

	if len(errors) > 0 {
		t.Fatalf("locking errors: %v", errors[0])
	}
	if lockingContext.IsLocked("browser") {
		t.Fatal("expected browser lock to be released")
	}
}

func TestQueueLockingReleasesOnError(t *testing.T) {
	lockingContext := NewLockManager()
	logger := &testLogger{}

	queue := NewQueue(QueueOpts{
		Actions: []Action{
			lockingAction("browser", func(_ *Context) error {
				return ErrInvalidScope
			}),
		},
		Name:           "TestQueue",
		End:            func() {},
		LockingContext: lockingContext,
		Logger:         logger,
	})

	queue.Run(map[string]any{
		"logger": logger,
	})

	if lockingContext.IsLocked("browser") {
		t.Fatal("expected lock to be released after error")
	}
}

func TestContextPushInsertsActionsFirst(t *testing.T) {
	order := []string{}
	lockingContext := NewLockManager()

	queue := NewQueue(QueueOpts{
		Actions: []Action{
			anyAction(func(ctx *Context) error {
				order = append(order, "first")
				ctx.Push([]Action{
					anyAction(func(_ *Context) error { order = append(order, "pushed-1"); return nil }),
					anyAction(func(_ *Context) error { order = append(order, "pushed-2"); return nil }),
				})
				return nil
			}),
			anyAction(func(_ *Context) error { order = append(order, "second"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: lockingContext,
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{})

	expected := []string{"first", "pushed-1", "pushed-2", "second"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, order)
		}
	}
}

func TestContextExtendAndName(t *testing.T) {
	seen := []string{}
	lockingContext := NewLockManager()

	queue := NewQueue(QueueOpts{
		Actions: []Action{
			anyAction(func(ctx *Context) error {
				ctx.Extend(map[string]any{"tag": "alpha"})
				return nil
			}),
			anyAction(func(ctx *Context) error {
				tag := ctx.Data["tag"].(string)
				seen = append(seen, ctx.Name()+"-"+tag)
				return nil
			}),
		},
		Name:           "QueueAlpha",
		LockingContext: lockingContext,
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{})

	if len(seen) != 1 || seen[0] != "QueueAlpha-alpha" {
		t.Fatalf("unexpected seen: %v", seen)
	}
}

func TestContextAbortClearsRemaining(t *testing.T) {
	order := []string{}
	lockingContext := NewLockManager()

	queue := NewQueue(QueueOpts{
		Actions: []Action{
			anyAction(func(ctx *Context) error {
				order = append(order, "first")
				ctx.Abort()
				return nil
			}),
			anyAction(func(_ *Context) error { order = append(order, "second"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: lockingContext,
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{})

	if len(order) != 1 || order[0] != "first" {
		t.Fatalf("unexpected order: %v", order)
	}
}

func TestDefaultOnErrorAbortsAndLogs(t *testing.T) {
	order := []string{}
	logger := &testLogger{}
	lockingContext := NewLockManager()

	failAction := anyAction(func(_ *Context) error {
		return ErrInvalidScope
	})

	queue := NewQueue(QueueOpts{
		Actions: []Action{
			failAction,
			anyAction(func(_ *Context) error { order = append(order, "after"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: lockingContext,
		Logger:         logger,
	})

	queue.Run(map[string]any{
		"logger": logger,
	})

	if len(order) != 0 {
		t.Fatalf("expected queue to stop, got %v", order)
	}
	if logger.ErrorCount() != 1 {
		t.Fatalf("expected logger error to be called once, got %d", logger.ErrorCount())
	}
}

func TestWithErrorHandlerContinuesQueue(t *testing.T) {
	order := []string{}
	lockingContext := NewLockManager()

	queue := NewQueue(QueueOpts{
		Actions: []Action{
			WithErrorHandler(func(_ *Context) error {
				return ErrInvalidScope
			}, func(_ error, ctx *Context) {
				ctx.Extend(map[string]any{"recovered": true})
			}),
			anyAction(func(ctx *Context) error {
				if recovered, ok := ctx.Data["recovered"].(bool); ok && recovered {
					order = append(order, "after")
				}
				return nil
			}),
		},
		Name:           "TestQueue",
		LockingContext: lockingContext,
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{})

	if len(order) != 1 || order[0] != "after" {
		t.Fatalf("unexpected order: %v", order)
	}
}
