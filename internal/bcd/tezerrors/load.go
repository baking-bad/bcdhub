package tezerrors

var errorDescriptions map[string]Description

// Description -
type Description struct {
	Title       string `json:"title"`
	Description string `json:"descr"`
}

// LoadErrorDescriptions -
func LoadErrorDescriptions() (err error) {
	errorDescriptions = make(map[string]Description)

	return json.Unmarshal([]byte(errorsDescrs), &errorDescriptions)
}
