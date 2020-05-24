package storage

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

// FindBigMapPointers -
func FindBigMapPointers(m meta.Metadata, storage gjson.Result) (map[int64]string, error) {
	key := make(map[int64]string)
	keyInt := storage.Get("int")

	if keyInt.Exists() {
		key[keyInt.Int()] = "0"
		return key, nil
	}

	for k, v := range m {
		if v.Prim != consts.BIGMAP {
			continue
		}

		if err := setMapPtr(storage, k, key); err != nil {
			return nil, err
		}
	}
	return key, nil
}

func setMapPtr(storage gjson.Result, path string, m map[int64]string) error {
	var buf strings.Builder

	trimmed := strings.TrimPrefix(path, "0/")
	for _, s := range strings.Split(trimmed, "/") {
		switch s {
		case "l", "s":
			buf.WriteString("#.")
		case "k":
			buf.WriteString("#.args.0.")
		case "v":
			buf.WriteString("#.args.1.")
		case "o":
			buf.WriteString("args.0.")
		default:
			buf.WriteString("args.")
			buf.WriteString(s)
			buf.WriteString(".")
		}
	}
	buf.WriteString("int")

	ptr := storage.Get(buf.String())
	if !ptr.Exists() {
		return fmt.Errorf("Path %s is not pointer: %s", path, buf.String())
	}

	for _, p := range ptr.Array() {
		if _, ok := m[p.Int()]; ok {
			return fmt.Errorf("Pointer already exists: %d", p.Int())
		}
		m[p.Int()] = path
	}

	return nil
}
