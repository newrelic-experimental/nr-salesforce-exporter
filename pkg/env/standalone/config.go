package standalone

import (
	"sync"
	"sync/atomic"
)

type SharedConfig[T any] struct {
	workerConfig    T
	workerConfigMu  sync.Mutex
	workerIsRunning atomic.Bool
}

func (w *SharedConfig[T]) SetIsRunning() bool {
	return w.workerIsRunning.Swap(true)
}

func (w *SharedConfig[T]) SetConfig(config T) {
	w.workerConfigMu.Lock()
	w.workerConfig = config
	w.workerConfigMu.Unlock()
}

func (w *SharedConfig[T]) Config() T {
	w.workerConfigMu.Lock()
	config := w.workerConfig
	w.workerConfigMu.Unlock()
	return config
}
