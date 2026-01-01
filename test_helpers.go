package queuerunner

import "sync"

type testLogger struct {
	mu     sync.Mutex
	infos  []string
	errors []error
}

func (logger *testLogger) Info(message string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.infos = append(logger.infos, message)
}

func (logger *testLogger) SetContext(_ string) {}

func (logger *testLogger) Error(err error) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.errors = append(logger.errors, err)
}

func (logger *testLogger) ErrorCount() int {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	return len(logger.errors)
}
