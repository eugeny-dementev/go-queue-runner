package queuerunner

import "time"

func WithErrorHandler(action Action, handler ErrorHandler) Action {
	return func(ctx *Context) error {
		err := action(ctx)
		if err != nil && handler != nil {
			handler(err, ctx)
			return nil
		}
		return err
	}
}

func WithDelay(action Action, delay time.Duration) Action {
	return func(ctx *Context) error {
		if err := action(ctx); err != nil {
			return err
		}
		if delay > 0 {
			time.Sleep(delay)
		}
		return nil
	}
}

func WithLock(scope string, action Action) Action {
	if err := ValidateScope(scope); err != nil {
		return func(_ *Context) error { return err }
	}

	return func(ctx *Context) error {
		if ctx == nil || ctx.locking == nil {
			return action(ctx)
		}
		return ctx.locking.RunWithLock(scope, func() error {
			return action(ctx)
		})
	}
}
