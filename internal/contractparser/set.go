package contractparser

// Set -
type Set []string

// Append - append items to set
func (ptr *Set) Append(str ...string) {
	s := *ptr
	for j := range str {
		if str[j] == "" {
			continue
		}
		found := false

		for i := range s {
			if s[i] == str[j] {
				found = true
				break
			}
		}

		if !found {
			s = append(s, str[j])
		}
	}
	*ptr = s
}

// Len - returns length of set
func (ptr *Set) Len() int {
	return len(*ptr)
}

// Clear - clears set
func (ptr *Set) Clear() {
	*ptr = nil
}
