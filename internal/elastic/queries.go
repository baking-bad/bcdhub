package elastic

var (
	queryAll = map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 10000,
	}
)

var allFields = []string{
	"address^10", "manager^8", "delegate^6", "tags^4", "hardcoded", "annotations", "fail_strings", "entrypoints", "language",
}

var mapFields = map[string]string{
	"address":   "address^10",
	"manager":   "manager^8",
	"delegate":  "delegate^6",
	"tags":      "tags^4",
	"hardcoded": "hardcoded",
	"annots":    "annotations",
	"fail":      "fail_strings",
	"entry":     "entrypoints",
	"language":  "language",
}

var mapHighlights = map[string]interface{}{
	"address":      map[string]interface{}{},
	"manager":      map[string]interface{}{},
	"delegate":     map[string]interface{}{},
	"tags":         map[string]interface{}{},
	"hardcoded":    map[string]interface{}{},
	"annotations":  map[string]interface{}{},
	"fail_strings": map[string]interface{}{},
	"entrypoints":  map[string]interface{}{},
	"language":     map[string]interface{}{},
}

var supportedNetworks = map[string]struct{}{
	"mainnet":     struct{}{},
	"zeronet":     struct{}{},
	"babylonnet":  struct{}{},
	"carthagenet": struct{}{},
}
