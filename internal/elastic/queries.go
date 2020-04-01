package elastic

var allFields = []string{
	"alias^10",
	"tags^9",
	"entrypoints^8",
	"entrypoint^8",
	"fail_strings^7",
	"errors.with^7",
	"errors.id^6",
	"language^5",
	"annotations^4",
	"delegate^2",
	"hardcoded^2",
	"manager",
	"address",
	"hash",
}

var mapFields = map[string]string{
	"alias":       "alias^10",
	"tags":        "tags^9",
	"entry":       "entrypoints^8",
	"entrypoint":  "entrypoint^8",
	"fail":        "fail_strings^7",
	"errors.with": "errors.with^7",
	"errors.id":   "errors.id^6",
	"language":    "language^5",
	"annots":      "annotations^4",
	"delegate":    "delegate^2",
	"hardcoded":   "hardcoded^2",
	"manager":     "manager",
	"hash":        "hash",
	"address":     "address",
}

var mapHighlights = qItem{
	"alias":        qItem{},
	"address":      qItem{},
	"hash":         qItem{},
	"manager":      qItem{},
	"delegate":     qItem{},
	"tags":         qItem{},
	"hardcoded":    qItem{},
	"annotations":  qItem{},
	"fail_strings": qItem{},
	"entrypoints":  qItem{},
	"language":     qItem{},
	"errors.id":    qItem{},
	"errors.with":  qItem{},
	"entrypoint":   qItem{},
}

var searchableInidices = []string{
	DocContracts,
	DocOperations,
}
