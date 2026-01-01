package queuerunner

import (
	"errors"
	"strings"
	"sync"
)

var ErrInvalidScope = errors.New("lock scope must be a non-empty string")

func ValidateScope(scope string) error {
	if strings.TrimSpace(scope) == "" {
		return ErrInvalidScope
	}
	return nil
}

type LockingContext interface {
	IsLocked(scope string) bool
	Lock(scope string) error
	Unlock(scope string)
	Wait(scope string) error
	RunWithLock(scope string, fn func() error) error
}

type LockManager struct {
	mu     sync.Mutex
	scopes map[string]chan struct{}
}

func NewLockManager() *LockManager {
	return &LockManager{
		scopes: map[string]chan struct{}{},
	}
}

func (manager *LockManager) IsLocked(scope string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	_, ok := manager.scopes[scope]
	return ok
}

func (manager *LockManager) Lock(scope string) error {
	if err := ValidateScope(scope); err != nil {
		return err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	if _, ok := manager.scopes[scope]; ok {
		return errors.New("scope is already locked")
	}

	manager.scopes[scope] = make(chan struct{})
	return nil
}

func (manager *LockManager) Unlock(scope string) {
	manager.mu.Lock()
	ch, ok := manager.scopes[scope]
	if ok {
		delete(manager.scopes, scope)
	}
	manager.mu.Unlock()

	if ok {
		close(ch)
	}
}

func (manager *LockManager) Wait(scope string) error {
	if err := ValidateScope(scope); err != nil {
		return err
	}

	manager.mu.Lock()
	ch, ok := manager.scopes[scope]
	manager.mu.Unlock()

	if !ok {
		return nil
	}

	<-ch
	return nil
}

func (manager *LockManager) RunWithLock(scope string, fn func() error) error {
	if err := ValidateScope(scope); err != nil {
		return err
	}

	for {
		manager.mu.Lock()
		ch, ok := manager.scopes[scope]
		if !ok {
			ch = make(chan struct{})
			manager.scopes[scope] = ch
			manager.mu.Unlock()

			err := fn()

			manager.mu.Lock()
			delete(manager.scopes, scope)
			manager.mu.Unlock()

			close(ch)
			return err
		}
		manager.mu.Unlock()

		<-ch
	}
}
