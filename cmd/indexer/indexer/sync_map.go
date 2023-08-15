package indexer

import "sync"

// Map -
type Map[K comparable, V any] struct {
	m  map[K]V
	mx *sync.RWMutex
}

// NewMap -
func NewMap[K comparable, V any]() Map[K, V] {
	return Map[K, V]{
		m:  make(map[K]V),
		mx: new(sync.RWMutex),
	}
}

// Set -
func (m Map[K, V]) Set(key K, value V) {
	m.mx.Lock()
	m.m[key] = value
	m.mx.Unlock()
}

// Get -
func (m Map[K, V]) Get(key K) (V, bool) {
	m.mx.RLock()
	value, ok := m.m[key]
	m.mx.RUnlock()
	return value, ok
}

// Exists -
func (m Map[K, V]) Exists(key K) bool {
	m.mx.RLock()
	_, ok := m.m[key]
	m.mx.RUnlock()
	return ok
}

// Delete -
func (m Map[K, V]) Delete(key K) {
	m.mx.Lock()
	delete(m.m, key)
	m.mx.Unlock()
}
