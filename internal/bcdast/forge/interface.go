package forge

// Forger -
type Forger interface {
	Forge() ([]byte, error)
}

// Unforger -
type Unforger interface {
	Unforge(data []byte) (int, error)
}
