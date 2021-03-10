package storage

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Errors -
var (
	ErrInvalidPath          = errors.Errorf("Invalid path")
	ErrPathIsNotPointer     = errors.Errorf("Path is not pointer")
	ErrPointerAlreadyExists = errors.Errorf("Pointer already exists")
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

		k := checkOnOr(k, m)

		if err := setMapPtr(storage, k, key); err != nil {
			return nil, err
		}
	}
	return key, nil
}

func checkOnOr(path string, m meta.Metadata) string {
	var buf strings.Builder
	parts := strings.Split(path, "/")

	var nextOr bool
	for i := 0; i < len(parts); i++ {
		if i > 0 {
			buf.WriteByte('/')
		}
		subPath := strings.Join(parts[:i+1], "/")
		node, ok := m[subPath]
		if !ok {
			return path
		}

		if nextOr {
			buf.WriteString("0")
		} else {
			buf.WriteString(parts[i])
		}

		nextOr = node.Prim == consts.OR
	}

	return buf.String()
}

func getJSONPath(path string) string {
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
	resp := buf.String()
	return strings.TrimSuffix(resp, ".")
}

func setMapPtr(storage gjson.Result, path string, m map[int64]string) error {
	pathJSON := getJSONPath(path) + ".int"

	ptr := storage.Get(pathJSON)
	if !ptr.Exists() {
		return errors.Wrapf(ErrPathIsNotPointer, "path=%s buf=%s", path, pathJSON)
	}

	for _, p := range ptr.Array() {
		if _, ok := m[p.Int()]; ok {
			return errors.Wrapf(ErrPointerAlreadyExists, "ptr=%d", p.Int())
		}
		m[p.Int()] = path
	}

	return nil
}

var emptyBigMap = gjson.Parse(`[]`)

// EnrichEmptyPointers -
func EnrichEmptyPointers(metadata meta.Metadata, storage gjson.Result) (gjson.Result, error) {
	if node, ok := metadata["0"]; ok && node.Type == consts.BIGMAP && storage.Get("int").Exists() {
		return emptyBigMap, nil
	}

	for path, node := range metadata {
		if node.Type != consts.BIGMAP {
			continue
		}

		ptrs := make(map[int64]string)
		if err := setMapPtr(storage, path, ptrs); err != nil {
			if errors.Is(err, ErrPathIsNotPointer) {
				continue
			}
			return storage, err
		}

		for ptr, jsonPath := range ptrs {
			binPath := strings.TrimPrefix(jsonPath, "0/")
			p := newmiguel.GetGJSONPath(binPath)
			fullJSONPath, err := findPtrJSONPath(ptr, p, storage)
			if err != nil {
				return storage, err
			}

			s, err := sjson.Set(storage.String(), fullJSONPath, []interface{}{})
			if err != nil {
				return storage, err
			}
			storage = gjson.Parse(s)
		}
	}

	return storage, nil
}

func findPtrJSONPath(ptr int64, path string, storage gjson.Result) (string, error) {
	val := storage
	parts := strings.Split(path, ".")

	var newPath strings.Builder
	for i := range parts {
		if parts[i] == "#" && val.IsArray() {
			for idx, item := range val.Array() {
				if i == len(parts)-1 {
					if ptr != item.Get("int").Int() {
						continue
					}
					if newPath.Len() != 0 {
						newPath.WriteString(".")
					}
					fmt.Fprintf(&newPath, "%d", idx)
					return newPath.String(), nil
				}

				p := strings.Join(parts[i+1:], ".")
				np, err := findPtrJSONPath(ptr, p, item)
				if err != nil {
					continue
				}
				if np != "" {
					fmt.Fprintf(&newPath, ".%d.%s", idx, strings.TrimPrefix(np, "."))
					return newPath.String(), nil
				}
			}
			return "", ErrInvalidPath
		}

		buf := val.Get(parts[i])
		if !buf.IsArray() && !buf.IsObject() {
			return "", ErrInvalidPath
		}
		if i == len(parts)-1 {
			if buf.Get("int").Exists() {
				if ptr != buf.Get("int").Int() {
					return "", ErrInvalidPath
				}
				if newPath.Len() != 0 {
					newPath.WriteString(".")
				}
				newPath.WriteString(parts[i])
				return newPath.String(), nil
			}
			for j := 0; j < int(buf.Int()); j++ {
				var bufPath strings.Builder
				fmt.Fprintf(&bufPath, "%d", j)
				if i < len(parts)-1 {
					fmt.Fprintf(&bufPath, ".%s", strings.Join(parts[i+1:], "."))
				}
				p, err := findPtrJSONPath(ptr, bufPath.String(), val)
				if err != nil {
					return "", err
				}
				if p != "" {
					fmt.Fprintf(&newPath, ".%s", p)
					return newPath.String(), nil
				}
			}
		} else {
			if newPath.Len() != 0 {
				newPath.WriteString(".")
			}

			newPath.WriteString(parts[i])
			val = buf
		}

	}
	return newPath.String(), nil
}
