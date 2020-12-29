package search

// Context -
type Context struct {
	Text       string
	Indices    []string
	Fields     []string
	Highlights map[string]interface{}
	Offset     int64
}

// NewContext -
func NewContext() Context {
	return Context{
		Fields:     make([]string, 0),
		Indices:    make([]string, 0),
		Highlights: make(map[string]interface{}),
	}
}
