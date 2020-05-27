package handlers

type aliasRequest struct {
	Address string `form:"address" binding:"required,address"`
	Network string `form:"network" binding:"required,network"`
	Alias   string `form:"alias"  binding:"required"`
}

type getContractCodeRequest struct {
	Address  string `uri:"address" binding:"required,address"`
	Network  string `uri:"network" binding:"required,network"`
	Protocol string `form:"protocol,omitempty"`
	Level    int64  `form:"level,omitempty"`
}

// CodeDiffLeg -
type CodeDiffLeg struct {
	Address  string `json:"address" binding:"required,address"`
	Network  string `json:"network" binding:"required,network"`
	Protocol string `json:"protocol,omitempty"`
	Level    int64  `json:"level,omitempty"`
}

// CodeDiffRequest -
type CodeDiffRequest struct {
	Left  CodeDiffLeg `json:"left" binding:"required"`
	Right CodeDiffLeg `json:"right" binding:"required"`
}

type getContractRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`
}

type getBigMapRequest struct {
	Network string `uri:"network" binding:"required,network"`
	Ptr     int64  `uri:"ptr" binding:"min=0"`
}

type getBigMapByKeyHashRequest struct {
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
	Address          string `json:"address" binding:"required"`
	Network          string `json:"network" binding:"required"`
	Alias            string `json:"alias"`
	WatchSame        bool   `json:"watchSame"`
	WatchSimilar     bool   `json:"watchSimilar"`
	WatchDeployed    bool   `json:"watchDeployed"`
	WatchMigrations  bool   `json:"watchMigrations"`
	WatchDeployments bool   `json:"watchDeployments"`
	WatchCalls       bool   `json:"watchCalls"`
	WatchErrors      bool   `json:"watchErrors"`
}

type voteRequest struct {
	SourceAddress      string `json:"src" binding:"required,address"`
	SourceNetwork      string `json:"src_network" binding:"required,network"`
	DestinationAddress string `json:"dest" binding:"required,address"`
	DestinationNetwork string `json:"dest_network" binding:"required,network"`
	Vote               uint   `json:"vote" binding:"oneof=0 1"`
}

// OPGRequest -
type OPGRequest struct {
	Hash string `uri:"hash" binding:"required,opg" example:"ooy4c6G2BZzybYEY3vRQ7WXGL63tFmamTeGTHdjUxhd6ckbSNnb"`
}

// FormatterRequest -
type FormatterRequest struct {
	Inline   bool   `form:"inline"`
	LineSize int    `form:"lineSize"`
	Code     string `form:"code"`
}

type getByNetwork struct {
	Network string `uri:"network" binding:"required,network" example:"mainnet"`
}

type bigMapSearchRequest struct {
	Offset int64  `form:"offset" binding:"min=0"`
	Size   int64  `form:"size" binding:"min=0"`
	Search string `form:"q"`
}

type getEntrypointDataRequest struct {
	BinPath string                 `json:"bin_path" binding:"required"`
	Data    map[string]interface{} `json:"data" binding:"required"`
	Format  string                 `json:"format"`
}

type getSeriesRequest struct {
	Index  string `form:"index" binding:"oneof=contract operation" example:"contract"`
	Period string `form:"period" binding:"oneof=year month week day" example:"year"`
}

type getBySlugRequest struct {
	Slug string `uri:"slug"  binding:"required"`
}

type getOperationByIDRequest struct {
	ID string `uri:"id" binding:"required"`
}

type runCodeRequest struct {
	Data     map[string]interface{} `json:"data" binding:"required"`
	BinPath  string                 `json:"bin_path" binding:"required"`
	Amount   int64                  `json:"amount,omitempty"`
	GasLimit int64                  `json:"gas_limit,omitempty"`
	Source   string                 `json:"source,omitempty" binding:"omitempty,address"`
	Sender   string                 `json:"sender,omitempty" binding:"omitempty,address"`
}

type markReadRequest struct {
	Timestamp int64 `json:"timestamp"`
}
