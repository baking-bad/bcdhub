package contractparser

// Set -
type Set map[string]struct{}

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
