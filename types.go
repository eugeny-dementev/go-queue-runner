package queuerunner

type Action func(*Context) error

type ErrorHandler func(err error, ctx *Context)

type Branches struct {
	Then []Action
	Else []Action
}

type EndListener func(name string, size int)
