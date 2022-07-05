package transfer

// Repository -
type Repository interface {
	GetAll(level int64) ([]Transfer, error)
	GetTransfered(contract string, tokenID uint64) (result float64, err error)
	GetToken24HoursVolume(contract string, initiators, entrypoints []string, tokenID uint64) (float64, error)
	CalcBalances(contract string) ([]Balance, error)
}
