package tzkt

// Service -
type Service interface {
	GetMempool(address string) ([]MempoolOperation, error)
}
