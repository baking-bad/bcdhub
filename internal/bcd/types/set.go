package types

// Set -
type Set map[string]struct{}

// Add - add item to set
func (s Set) Add(item string) {
	s[item] = struct{}{}
}

// Append - append items to set
func (s Set) Append(str ...string) {
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
func (s Set) Len() int {
	return len(s)
}

// Values - return keys
func (s Set) Values() []string {
	r := make([]string, 0)
	for k := range s {
		r = append(r, k)
	}
	return r
}

// Merge -
func (s Set) Merge(m Set) {
	for k := range m {
		s[k] = struct{}{}
	}
}

// ArrayUniqueLen -
func ArrayUniqueLen(arr []string) int {
	buf := make(map[string]struct{})
	for i := range arr {
		buf[arr[i]] = struct{}{}
	}
	return len(buf)
}
