package queuerunner

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type RunnerOpts struct {
	Logger Logger
}

type QueueRunner struct {
	queues    map[string]*Queue
	listeners []EndListener
	logger    Logger
	locking   LockingContext
	mu        sync.Mutex
	counter   uint64
}

func NewQueueRunner(opts RunnerOpts) *QueueRunner {
	runner := &QueueRunner{
		queues:  map[string]*Queue{},
		locking: NewLockManager(),
	}

	if opts.Logger != nil {
		runner.logger = opts.Logger
	}

	return runner
}

func (runner *QueueRunner) PreparteLockingContext() LockingContext {
	return runner.locking
}

func (runner *QueueRunner) PrepareLockingContext() LockingContext {
	return runner.locking
}

func (runner *QueueRunner) Add(actions []Action, context map[string]any, name string) {
	queueName := name
	if queueName == "" {
		queueName = runner.getName()
	}

	queue := NewQueue(QueueOpts{
		Name:           queueName,
		Actions:        actions,
		End:            func() { runner.onQueueEnd(queueName) },
		Logger:         runner.logger,
		LockingContext: runner.locking,
	})

	runner.mu.Lock()
	runner.queues[queueName] = queue
	runner.mu.Unlock()

	go queue.Run(context)
}

func (runner *QueueRunner) AddEndListener(listener EndListener) {
	runner.listeners = append(runner.listeners, listener)
}

func (runner *QueueRunner) onQueueEnd(name string) {
	runner.mu.Lock()
	delete(runner.queues, name)
	size := len(runner.queues)
	runner.mu.Unlock()

	for _, listener := range runner.listeners {
		listener(name, size)
	}
}

func (runner *QueueRunner) getName() string {
	id := atomic.AddUint64(&runner.counter, 1)
	return fmt.Sprintf("queue-%d", id)
}
