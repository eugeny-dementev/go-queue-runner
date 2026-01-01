package queuerunner

import "testing"

func TestUtilIfThenBranch(t *testing.T) {
	order := []string{}
	queue := NewQueue(QueueOpts{
		Actions: []Action{
			Util.If(func(ctx *Context) (bool, error) {
				return ctx.Data["flag"].(bool), nil
			}, Branches{
				Then: []Action{
					anyAction(func(_ *Context) error { order = append(order, "then-1"); return nil }),
					anyAction(func(_ *Context) error { order = append(order, "then-2"); return nil }),
				},
				Else: []Action{
					anyAction(func(_ *Context) error { order = append(order, "else"); return nil }),
				},
			}),
			anyAction(func(_ *Context) error { order = append(order, "after"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: NewLockManager(),
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{"flag": true})

	expected := []string{"then-1", "then-2", "after"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestUtilIfElseBranch(t *testing.T) {
	order := []string{}
	queue := NewQueue(QueueOpts{
		Actions: []Action{
			Util.If(func(ctx *Context) (bool, error) {
				return ctx.Data["flag"].(bool), nil
			}, Branches{
				Then: []Action{
					anyAction(func(_ *Context) error { order = append(order, "then"); return nil }),
				},
				Else: []Action{
					anyAction(func(_ *Context) error { order = append(order, "else"); return nil }),
				},
			}),
			anyAction(func(_ *Context) error { order = append(order, "after"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: NewLockManager(),
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{"flag": false})

	expected := []string{"else", "after"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestUtilIfWithoutElse(t *testing.T) {
	order := []string{}
	queue := NewQueue(QueueOpts{
		Actions: []Action{
			Util.If(func(ctx *Context) (bool, error) {
				return ctx.Data["flag"].(bool), nil
			}, Branches{
				Then: []Action{
					anyAction(func(_ *Context) error { order = append(order, "then"); return nil }),
				},
			}),
			anyAction(func(_ *Context) error { order = append(order, "after"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: NewLockManager(),
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{"flag": false})

	expected := []string{"after"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestUtilValid(t *testing.T) {
	order := []string{}
	queue := NewQueue(QueueOpts{
		Actions: []Action{
			Util.Valid(func(ctx *Context) (bool, error) {
				return ctx.Data["count"].(int) > 0, nil
			}, []Action{
				anyAction(func(_ *Context) error { order = append(order, "valid"); return nil }),
			}),
			anyAction(func(_ *Context) error { order = append(order, "after"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: NewLockManager(),
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{"count": 1})

	expected := []string{"valid", "after"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestUtilValidSkips(t *testing.T) {
	order := []string{}
	queue := NewQueue(QueueOpts{
		Actions: []Action{
			Util.Valid(func(ctx *Context) (bool, error) {
				return ctx.Data["count"].(int) > 0, nil
			}, []Action{
				anyAction(func(_ *Context) error { order = append(order, "valid"); return nil }),
			}),
			anyAction(func(_ *Context) error { order = append(order, "after"); return nil }),
		},
		Name:           "TestQueue",
		LockingContext: NewLockManager(),
		Logger:         &testLogger{},
	})

	queue.Run(map[string]any{"count": 0})

	expected := []string{"after"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}
