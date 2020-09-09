package models

import "github.com/tidwall/gjson"

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

// ParseElasticJSON -
func (p *Protocol) ParseElasticJSON(hit gjson.Result) {
	p.ID = hit.Get("_id").String()
	p.Hash = hit.Get("_source.hash").String()
	p.Network = hit.Get("_source.network").String()
	p.StartLevel = hit.Get("_source.start_level").Int()
	p.EndLevel = hit.Get("_source.end_level").Int()
	p.Alias = hit.Get("_source.alias").String()
	p.SymLink = hit.Get("_source.sym_link").String()
	p.Constants = Constants{
		CostPerByte:                  hit.Get("_source.constants.cost_per_byte").Int(),
		HardGasLimitPerOperation:     hit.Get("_source.constants.hard_gas_limit_per_operation").Int(),
		HardStorageLimitPerOperation: hit.Get("_source.constants.hard_storage_limit_per_operation").Int(),
		TimeBetweenBlocks:            hit.Get("_source.constants.time_between_blocks").Int(),
	}
}

// GetQueue -
func (p *Protocol) GetQueue() string {
	return ""
}
