package protocol

// Protocol -
type Protocol struct {
	ID int64 `json:"-"`

	Hash       string `json:"hash"`
	Network    string `json:"network"`
	StartLevel int64  `json:"start_level" gorm:",default:0"`
	EndLevel   int64  `json:"end_level" gorm:",default:0"`
	SymLink    string `json:"sym_link"`
	Alias      string `json:"alias"`
	*Constants
}

// Constants -
type Constants struct {
	CostPerByte                  int64 `json:"cost_per_byte" gorm:",default:0"`
	HardGasLimitPerOperation     int64 `json:"hard_gas_limit_per_operation" gorm:",default:0"`
	HardStorageLimitPerOperation int64 `json:"hard_storage_limit_per_operation" gorm:",default:0"`
	TimeBetweenBlocks            int64 `json:"time_between_blocks" gorm:",default:0"`
}

// GetID -
func (p *Protocol) GetID() int64 {
	return p.ID
}

// GetIndex -
func (p *Protocol) GetIndex() string {
	return "protocols"
}

// MarshalToQueue -
func (p *Protocol) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// GetQueues -
func (p *Protocol) GetQueues() []string {
	return nil
}
