package queuerunner

type Context struct {
	Data   map[string]any
	Logger Logger

	pushFn  func(actions []Action)
	nameFn  func() string
	abortFn func()
	locking LockingContext
}

func newContext(pushFn func([]Action), nameFn func() string, abortFn func(), locking LockingContext) *Context {
	return &Context{
		Data:    map[string]any{},
		pushFn:  pushFn,
		nameFn:  nameFn,
		abortFn: abortFn,
		locking: locking,
	}
}

func (ctx *Context) Initialize(initial map[string]any) {
	ctx.Data = map[string]any{}
	for key, value := range initial {
		ctx.Data[key] = value
		if key == "logger" {
			if logger, ok := value.(Logger); ok {
				ctx.Logger = logger
			}
		}
	}
}

func (ctx *Context) Extend(values map[string]any) {
	if ctx.Data == nil {
		ctx.Data = map[string]any{}
	}
	for key, value := range values {
		ctx.Data[key] = value
		if key == "logger" {
			if logger, ok := value.(Logger); ok {
				ctx.Logger = logger
			}
		}
	}
}

func (ctx *Context) Push(actions []Action) {
	if ctx.pushFn == nil {
		return
	}
	ctx.pushFn(actions)
}

func (ctx *Context) Get(key string) (any, bool) {
	if ctx.Data == nil {
		return nil, false
	}
	value, ok := ctx.Data[key]
	return value, ok
}

func (ctx *Context) Set(key string, value any) {
	if ctx.Data == nil {
		ctx.Data = map[string]any{}
	}
	ctx.Data[key] = value
}

func (ctx *Context) Name() string {
	if ctx.nameFn == nil {
		return ""
	}
	return ctx.nameFn()
}

func (ctx *Context) Abort() {
	if ctx.abortFn == nil {
		return
	}
	ctx.abortFn()
}

func (ctx *Context) LoggerFromData() Logger {
	if ctx.Data == nil {
		return nil
	}

	if logger, ok := ctx.Data["logger"].(Logger); ok {
		return logger
	}

	return nil
}
