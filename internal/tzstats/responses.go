package tzstats

// ErrorResponse - All error messages are JSON encoded. They contain fields numeric and human readable fields to help developers easily debug and map errors.
type ErrorResponse struct {
	Errors []struct {
		Code      int    `json:"code"`
		Status    int    `json:"status"`
		Message   string `json:"message"`
		Scope     string `json:"scope"`
		Detail    string `json:"detail"`
		RequestID string `json:"request_id"`
	} `json:"errors"`
}

// Row - JSON row
type Row []interface{}

// TableResponse - default table response
type TableResponse []Row
