package protocol

// Repository -
type Repository interface {
	GetProtocol(string, string, int64) (Protocol, error)
	GetSymLinks(string, int64) (map[string]struct{}, error)
}
