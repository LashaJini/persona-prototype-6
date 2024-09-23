package safemap

import "sync"

type SafeMap[V interface{}] struct {
	Items map[string]V
	sync.RWMutex
}

func NewSafeMap[V interface{}]() *SafeMap[V] {
	items := make(map[string]V)
	sm := &SafeMap[V]{
		Items: items,
	}

	return sm
}
