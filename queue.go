package queuerunner

import (
	"fmt"
	"reflect"
	"runtime"
)

type QueueOpts struct {
	Actions        []Action
	Name           string
	End            func()
	Logger         Logger
	LockingContext LockingContext
	OnError        ErrorHandler
}

type Queue struct {
	name        string
	queue       []Action
	end         func()
	logger      Logger
	lockManager LockingContext
	context     *Context
	onError     ErrorHandler
}

func NewQueue(opts QueueOpts) *Queue {
	queueName := opts.Name
	if queueName == "" {
		queueName = "default queue name"
	}

	queue := &Queue{
		name:        queueName,
		queue:       append([]Action{}, opts.Actions...),
		end:         opts.End,
		logger:      opts.Logger,
		lockManager: opts.LockingContext,
		onError:     opts.OnError,
	}

	if queue.end == nil {
		queue.end = func() {}
	}
	if queue.logger == nil {
		queue.logger = defaultLogger()
	}
	if queue.lockManager == nil {
		queue.lockManager = NewLockManager()
	}
	if queue.onError == nil {
		queue.onError = defaultErrorHandler
	}

	queue.context = newContext(queue.Push, func() string { return queue.name }, queue.Abort, queue.lockManager)

	return queue
}

func (queue *Queue) Run(initial map[string]any) {
	if initial == nil {
		initial = map[string]any{}
	}

	queue.context.Initialize(initial)

	defer func() {
		queue.end()
	}()

	defer func() {
		if recovered := recover(); recovered != nil {
			queue.logger.Info(fmt.Sprintf("Queue(%s) failed", queue.name))
			if err, ok := recovered.(error); ok {
				queue.logger.Error(err)
			} else {
				queue.logger.Error(fmt.Errorf("%v", recovered))
			}
		}
	}()

	for {
		if len(queue.queue) == 0 {
			queue.logger.Info(fmt.Sprintf("Queue(%s): stopped", queue.name))
			break
		}

		action := queue.queue[0]
		queue.queue = queue.queue[1:]

		queue.logger.SetContext(actionName(action))
		queue.logger.Info(fmt.Sprintf("Queue(%s): running action", queue.name))

		if err := queue.executeSafe(action); err != nil {
			queue.handleError(err)
		}
	}
}

func (queue *Queue) executeSafe(action Action) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if recoveredErr, ok := recovered.(error); ok {
				err = recoveredErr
			} else {
				err = fmt.Errorf("%v", recovered)
			}
		}
	}()

	return action(queue.context)
}

func (queue *Queue) handleError(err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			queue.logger.Info(fmt.Sprintf("Queue(%s) onError failed", queue.name))
			if recoveredErr, ok := recovered.(error); ok {
				queue.logger.Error(recoveredErr)
			} else {
				queue.logger.Error(fmt.Errorf("%v", recovered))
			}
			queue.context.Abort()
		}
	}()

	queue.onError(err, queue.context)
}

func (queue *Queue) Push(actions []Action) {
	if len(actions) == 0 {
		return
	}

	queue.queue = append(actions, queue.queue...)
}

func (queue *Queue) Abort() {
	queue.queue = queue.queue[:0]
}

func actionName(action Action) string {
	if action == nil {
		return "unknown"
	}

	fn := runtime.FuncForPC(reflect.ValueOf(action).Pointer())
	if fn == nil {
		return "action"
	}

	return fn.Name()
}

func defaultErrorHandler(err error, ctx *Context) {
	if ctx == nil {
		return
	}

	logger := ctx.Logger
	if logger == nil {
		logger = ctx.LoggerFromData()
	}

	if logger != nil {
		logger.Error(err)
	}

	ctx.Abort()
}
