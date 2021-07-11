package service

// Repository -
type Repository interface {
	All() (state []State, err error)
	Get(name string) (State, error)
	Save(s State) error
}
