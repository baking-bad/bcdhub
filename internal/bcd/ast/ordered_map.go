package ast

import "sync"

// OrderedMap -
type OrderedMap struct {
	keys   []Comparable
	values map[Comparable]Comparable

	lock sync.Mutex
}

// NewOrderedMap -
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		keys:   make([]Comparable, 0),
		values: make(map[Comparable]Comparable),
	}
}

// Get -
func (m *OrderedMap) Get(key Comparable) (Comparable, bool) {
	defer m.lock.Unlock()
	m.lock.Lock()

	for k, v := range m.values {
		ok, err := key.Compare(k)
		if err != nil {
			return nil, false
		}
		if ok != 0 {
			continue
		}
		return v, true
	}
	return nil, false
}

func (m *OrderedMap) set(key, value Comparable) bool {
	defer m.lock.Unlock()
	m.lock.Lock()

	for k := range m.values {
		ok, err := key.Compare(k)
		if err != nil {
			return false
		}
		if ok != 0 {
			continue
		}
		m.values[k] = value
		return true
	}
	return false
}

// Add -
func (m *OrderedMap) Add(key, value Comparable) error {
	if ok := m.set(key, value); ok {
		return nil
	}

	defer m.lock.Unlock()
	m.lock.Lock()
	m.values[key] = value
	idx := -1
	for i := range m.keys {
		val, err := m.keys[i].Compare(key)
		if err != nil {
			return err
		}
		if val == 1 {
			idx = i
			break
		}
	}

	if idx == -1 {
		m.keys = append(m.keys, key)
	} else {
		m.keys = append(m.keys[:idx+1], m.keys[idx:]...)
		m.keys[idx] = key
	}
	return nil
}

// Remove -
func (m *OrderedMap) Remove(key Comparable) (Comparable, bool) {
	val, ok := m.Get(key)
	if !ok {
		return nil, false
	}

	defer m.lock.Unlock()
	m.lock.Lock()

	for i := range m.keys {
		res, err := m.keys[i].Compare(key)
		if err != nil {
			return nil, false
		}
		if res == 0 {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			delete(m.values, m.keys[i])
			return val, true
		}
	}
	return nil, false
}

// Range -
func (m *OrderedMap) Range(handler func(key, value Comparable) (bool, error)) error {
	defer m.lock.Unlock()
	m.lock.Lock()

	for i := range m.keys {
		isBreak, err := handler(m.keys[i], m.values[m.keys[i]])
		if err != nil || isBreak {
			return err
		}
	}
	return nil
}

// Len -
func (m *OrderedMap) Len() int {
	return len(m.keys)
}
