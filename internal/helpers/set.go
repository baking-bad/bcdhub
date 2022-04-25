package helpers

// SetKey -
type SetKey interface {
	~string
}

// Set -
type Set[T SetKey] map[T]struct{}

// Add - add item to set
func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

// Append - append items to set
func (s Set[T]) Append(str ...T) {
	for j := range str {
		if str[j] == "" {
			continue
		}

		if _, ok := s[str[j]]; !ok {
			s[str[j]] = struct{}{}
		}
	}
}

// Len - returns length of set
func (s Set[T]) Len() int {
	return len(s)
}

// Values - return keys
func (s Set[T]) Values() []T {
	r := make([]T, 0)
	for k := range s {
		r = append(r, k)
	}
	return r
}

// Merge -
func (s Set[T]) Merge(m Set[T]) {
	for k := range m {
		s[k] = struct{}{}
	}
}
