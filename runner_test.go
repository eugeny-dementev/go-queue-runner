package queuerunner

import (
	"sync"
	"testing"
	"time"
)

func TestRunnerOrder(t *testing.T) {
	var mu sync.Mutex
	order := []string{}
	done := make(chan struct{})

	logger := &testLogger{}
	runner := NewQueueRunner(RunnerOpts{Logger: logger})

	runner.AddEndListener(func(_ string, size int) {
		if size == 0 {
			close(done)
		}
	})

	queue1 := []Action{
		Util.Delay(10 * time.Millisecond),
		anyAction(func(_ *Context) error {
			mu.Lock()
			order = append(order, "second")
			mu.Unlock()
			return nil
		}),
	}
	queue2 := []Action{
		anyAction(func(_ *Context) error {
			mu.Lock()
			order = append(order, "first")
			mu.Unlock()
			return nil
		}),
		Util.Delay(15 * time.Millisecond),
		anyAction(func(_ *Context) error {
			mu.Lock()
			order = append(order, "third")
			mu.Unlock()
			return nil
		}),
	}

	runner.Add(queue1, map[string]any{}, "")
	runner.Add(queue2, map[string]any{}, "")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for queues")
	}

	expected := []string{"first", "second", "third"}
	if len(order) != len(expected) {
		t.Fatalf("unexpected order: %v", order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestRunnerLockingContext(t *testing.T) {
	runner := NewQueueRunner(RunnerOpts{})
	locking := runner.PreparteLockingContext()

	if err := locking.Lock("browser"); err != nil {
		t.Fatalf("unexpected lock error: %v", err)
	}

	if !locking.IsLocked("browser") {
		t.Fatal("expected browser scope to be locked")
	}

	done := make(chan struct{})
	go func() {
		_ = locking.Wait("browser")
		close(done)
	}()

	time.Sleep(5 * time.Millisecond)
	locking.Unlock("browser")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for lock release")
	}
}

func TestRunnerAbortAction(t *testing.T) {
	var mu sync.Mutex
	order := []string{}
	done := make(chan struct{})

	runner := NewQueueRunner(RunnerOpts{Logger: &testLogger{}})
	runner.AddEndListener(func(_ string, size int) {
		if size == 0 {
			close(done)
		}
	})

	queue := []Action{
		anyAction(func(_ *Context) error {
			mu.Lock()
			order = append(order, "first")
			mu.Unlock()
			return nil
		}),
		anyAction(func(_ *Context) error {
			mu.Lock()
			order = append(order, "second")
			mu.Unlock()
			return nil
		}),
		Util.Abort,
		anyAction(func(_ *Context) error {
			mu.Lock()
			order = append(order, "third")
			mu.Unlock()
			return nil
		}),
	}

	runner.Add(queue, map[string]any{}, "")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for queue")
	}

	expected := []string{"first", "second"}
	if len(order) != len(expected) {
		t.Fatalf("unexpected order: %v", order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestRunnerAddUsesNameAndContext(t *testing.T) {
	var mu sync.Mutex
	seen := []string{}
	done := make(chan struct{})

	runner := NewQueueRunner(RunnerOpts{Logger: &testLogger{}})
	runner.AddEndListener(func(name string, size int) {
		if size != 0 {
			return
		}
		if name != "named-queue" {
			t.Fatalf("expected named-queue, got %q", name)
		}
		close(done)
	})

	queue := []Action{
		anyAction(func(ctx *Context) error {
			mu.Lock()
			seen = append(seen, ctx.Name()+"-"+ctx.Data["tag"].(string))
			mu.Unlock()
			return nil
		}),
	}

	runner.Add(queue, map[string]any{"tag": "ok"}, "named-queue")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for queue")
	}

	if len(seen) != 1 || seen[0] != "named-queue-ok" {
		t.Fatalf("unexpected seen: %v", seen)
	}
}
