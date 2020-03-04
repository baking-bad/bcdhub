package handlers

type aliasRequest struct {
	Address string `form:"address" binding:"required,address"`
	Network string `form:"network" binding:"required,network"`
	Alias   string `form:"alias"  binding:"required"`
}

type getContractCodeRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`

	Level int64 `form:"level,omitempty"`
}

type getDiffRequest struct {
	SourceAddress      string `form:"sa" binding:"required,address"`
	SourceNetwork      string `form:"sn" binding:"required,network"`
	DestinationAddress string `form:"da" binding:"required,address"`
	DestinationNetwork string `form:"dn" binding:"required,network"`
}

type getContractRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`
}

// OauthRequest -
type OauthRequest struct {
	State string `form:"state"`
	Code  string `form:"code"`
}

// OauthParams -
type OauthParams struct {
	Provider string `uri:"provider"`
}

type offsetRequest struct {
	LastID string `form:"last_id" binding:"omitempty,numeric"`
}

type pageableRequest struct {
	Offset int64 `form:"offset" binding:"min=0"`
}

type searchRequest struct {
	Text     string `form:"q"`
	Fields   string `form:"f,omitempty"`
	Networks string `form:"n,omitempty"`
	Offset   uint   `form:"o,omitempty"`
	DateFrom uint   `form:"s,omitempty"`
	DateTo   uint   `form:"e,omitempty"`
	Grouping uint   `form:"g,omitempty"`
}

type subRequest struct {
	ID   string `json:"id" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type voteRequest struct {
	SourceAddress      string `json:"src" binding:"required,address"`
	SourceNetwork      string `json:"src_network" binding:"required,network"`
	DestinationAddress string `json:"dest" binding:"required,address"`
	DestinationNetwork string `json:"dest_network" binding:"required,network"`
	Vote               int    `json:"vote" binding:"required,oneof=0 1"`
}

// OPGRequest -
type OPGRequest struct {
	Hash string `uri:"hash" binding:"required,opg"`
}
