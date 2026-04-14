package platform

import (
	"context"
	"sync"
)

type Hook func(context.Context) error

type Lifecycle struct {
	mu         sync.RWMutex
	onStart    []Hook
	onShutdown []Hook
}

func NewLifecycle() *Lifecycle {
	return &Lifecycle{}
}

func (lc *Lifecycle) RegisterOnStart(hook Hook) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.onStart = append(lc.onStart, hook)
}

func (lc *Lifecycle) RegisterOnShutdown(hook Hook) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.onShutdown = append(lc.onShutdown, hook)
}

func (lc *Lifecycle) Start(ctx context.Context) error {
	lc.mu.RLock()
	hooks := append([]Hook{}, lc.onStart...)
	lc.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (lc *Lifecycle) Shutdown(ctx context.Context) error {
	lc.mu.RLock()
	hooks := append([]Hook{}, lc.onShutdown...)
	lc.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}
