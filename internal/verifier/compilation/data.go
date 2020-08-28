package compilation

// Task -
type Task struct {
	ID    uint
	Kind  string
	Files []string
	Dir   string
}
