package protocol

// Protocol -
type Protocol struct {
	ID string `json:"-"`

	Hash       string    `json:"hash"`
	Network    string    `json:"network"`
	StartLevel int64     `json:"start_level"`
	EndLevel   int64     `json:"end_level"`
	SymLink    string    `json:"sym_link"`
	Alias      string    `json:"alias"`
	Constants  Constants `json:"constants"`
}

// Constants -
type Constants struct {
	CostPerByte                  int64 `json:"cost_per_byte"`
	HardGasLimitPerOperation     int64 `json:"hard_gas_limit_per_operation"`
	HardStorageLimitPerOperation int64 `json:"hard_storage_limit_per_operation"`
	TimeBetweenBlocks            int64 `json:"time_between_blocks"`
}

// GetID -
func (p *Protocol) GetID() string {
	return p.ID
}

// GetIndex -
func (p *Protocol) GetIndex() string {
	return "protocol"
}

// MarshalToQueue -
func (p *Protocol) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// GetQueues -
func (p *Protocol) GetQueues() []string {
	return nil
}
