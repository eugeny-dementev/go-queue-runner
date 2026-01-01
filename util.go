package queuerunner

import "time"

type Condition func(ctx *Context) (bool, error)

type Validator func(ctx *Context) (bool, error)

type utilHelper struct {
	Abort Action
}

var Util = utilHelper{
	Abort: func(ctx *Context) error {
		if ctx != nil {
			ctx.Abort()
		}
		return nil
	},
}

func (utilHelper) Delay(timeout time.Duration) Action {
	return func(_ *Context) error {
		if timeout > 0 {
			time.Sleep(timeout)
		}
		return nil
	}
}

func (utilHelper) If(condition Condition, branches Branches) Action {
	return func(ctx *Context) error {
		result, err := condition(ctx)
		if err != nil {
			return err
		}

		if result {
			ctx.Push(branches.Then)
		} else if len(branches.Else) > 0 {
			ctx.Push(branches.Else)
		}

		return nil
	}
}

func (utilHelper) Valid(validator Validator, actions []Action) Action {
	return func(ctx *Context) error {
		result, err := validator(ctx)
		if err != nil {
			return err
		}

		if result {
			ctx.Push(actions)
		}

		return nil
	}
}
