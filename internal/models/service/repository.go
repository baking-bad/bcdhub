package service

// Repository -
type Repository interface {
	Get(name string) (State, error)
	Save(s State) error
}
