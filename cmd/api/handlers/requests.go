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

type getBigMapRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`
	Ptr     int64  `uri:"ptr" binding:"min=0"`
}

type getBigMapByKeyHashRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`
	Ptr     int64  `uri:"ptr" binding:"min=0"`
	KeyHash string `uri:"key_hash" binding:"required"`
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

type operationsRequest struct {
	LastID      string `form:"last_id" binding:"omitempty,numeric"`
	From        uint   `form:"from" binding:"omitempty"`
	To          uint   `form:"to" binding:"omitempty,gtfield=From"`
	Size        uint64 `form:"size" binding:"min=0"`
	Status      string `form:"status" binding:"omitempty,status"`
	Entrypoints string `form:"entrypoints" binding:"omitempty,excludesall=\"'"`
}

type pageableRequest struct {
	Offset int64 `form:"offset" binding:"min=0"`
	Size   int64 `form:"size" binding:"min=0"`
}

type cursorRequest struct {
	LastID string `form:"last_id" binding:"omitempty,numeric"`
	Size   int64  `form:"size" binding:"min=0"`
}

type searchRequest struct {
	Text      string `form:"q"`
	Fields    string `form:"f,omitempty"`
	Networks  string `form:"n,omitempty"`
	Offset    uint   `form:"o,omitempty"`
	DateFrom  uint   `form:"s,omitempty"`
	DateTo    uint   `form:"e,omitempty"`
	Grouping  uint   `form:"g,omitempty"`
	Indices   string `form:"i,omitempty"`
	Languages string `form:"l,omitempty"`
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
	Vote               int    `json:"vote" binding:"oneof=0 1"`
}

// OPGRequest -
type OPGRequest struct {
	Hash string `uri:"hash" binding:"required,opg"`
}

// FormatterRequest -
type FormatterRequest struct {
	Inline   bool   `form:"inline"`
	LineSize int    `form:"lineSize"`
	Code     string `form:"code"`
}

type getMigrationRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`

	Protocol string `form:"protocol"`
}

type getByNetwork struct {
	Network string `uri:"network" binding:"required,network"`
}

type bigMapSearchRequest struct {
	Offset int64  `form:"offset" binding:"min=0"`
	Size   int64  `form:"size" binding:"min=0"`
	Search string `form:"q"`
}
