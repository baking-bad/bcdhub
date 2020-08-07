package jsonschema

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/tidwall/gjson"
)

type maker interface {
	Do(string, meta.Metadata) (Schema, error)
}

// Schema -
type Schema map[string]interface{}

// DefaultModel -
type DefaultModel map[string]interface{}

// Extend -
func (model DefaultModel) Extend(another DefaultModel, binPath string) {
	if !strings.HasSuffix(binPath, "/o") {
		for k, v := range another {
			model[k] = v
		}
	} else {
		optionMap := make(DefaultModel)
		for k, v := range another {
			optionMap[k] = v
		}
		model[binPath] = optionMap
	}
}

// Fill - fill `model` from `data` by `metadata`
func (model DefaultModel) Fill(data gjson.Result, metadata meta.Metadata) error {
	root, ok := metadata["0"]
	if !ok {
		return fmt.Errorf("I've got no roots, but my home was never on the ground")
	}
	for k := range model {
		delete(model, k)
	}
	return model.fill(data, metadata, root, "0", "0", false)
}

// FillForEntrypoint - fill `model` from `data` by entrypoint `metadata`
func (model DefaultModel) FillForEntrypoint(data gjson.Result, metadata meta.Metadata, entrypoint string) error {
	root, ok := metadata["0"]
	if !ok {
		return fmt.Errorf("I've got no roots, but my home was never on the ground")
	}
	for k := range model {
		delete(model, k)
	}
	path := "0"

	for i := range root.Args {
		nm, ok := metadata[root.Args[i]]
		if !ok {
			return fmt.Errorf("Unknown node: %s", root.Args[i])
		}
		if nm.Name == entrypoint {
			path = root.Args[i]
			root = nm
			prim := data.Get("prim|@lower").String()
			for prim == consts.LEFT || prim == consts.RIGHT {
				data = data.Get("args.0")
				prim = data.Get("prim|@lower").String()
			}
			break
		}
	}

	return model.fill(data, metadata, root, path, path, false)
}

func (model DefaultModel) fill(data gjson.Result, metadata meta.Metadata, node *meta.NodeMetadata, path, prefix string, isOption bool, indices ...int) error {
	if !isOption {
		if err := model.optionWrapper(data, metadata, node, path, prefix, indices...); err != nil {
			return err
		}
		if _, ok := model[strings.TrimSuffix(path, "/o")]; ok {
			return nil
		}
	}

	binPath := prepareBinPath(path, prefix)

	switch node.Prim {
	case consts.PAIR:
		for _, argPath := range node.Args {
			arg, ok := metadata[argPath]
			if !ok {
				return fmt.Errorf("Unknown pair arg path: %s", argPath)
			}

			if err := model.fill(data, metadata, arg, argPath, prefix, false, indices...); err != nil {
				return err
			}
		}
	case consts.LIST, consts.SET:
		suffix := "l"
		if node.Prim == consts.SET {
			suffix = "s"
		}

		listPath := fmt.Sprintf("%s/%s", path, suffix)
		jsonPath := getGJSONPath(path, binPath, indices...)
		arr := data.Get(jsonPath).Array()
		itemNode, ok := metadata[listPath]
		if !ok {
			return fmt.Errorf("Unknown list node: %s", listPath)
		}
		result := make([]interface{}, 0)
		for i := range arr {
			itemModel := make(DefaultModel)
			newIndices := append(indices, i)
			if err := itemModel.fill(data, metadata, itemNode, listPath, prefix, false, newIndices...); err != nil {
				return err
			}
			result = append(result, itemModel)
		}
		model[path] = result
	case consts.INT, consts.NAT, consts.MUTEZ:
		jsonPath := getGJSONPath(path, binPath, indices...)
		i := getDataByPath(jsonPath, consts.INT, data)
		model[path] = i.Int()
	case consts.STRING, consts.KEY, consts.KEYHASH, consts.CONTRACT, consts.ADDRESS, consts.SIGNATURE:
		jsonPath := getGJSONPath(path, binPath, indices...)

		str := getDataByPath(jsonPath, consts.STRING, data)
		s := str.String()
		if !str.Exists() {
			str = getDataByPath(jsonPath, consts.BYTES, data)
			s = unpack.Bytes(str.String())
		}

		model[path] = s
	case consts.TIMESTAMP:
		jsonPath := getGJSONPath(path, binPath, indices...)

		str := getDataByPath(jsonPath, consts.STRING, data)
		s := str.String()
		if !str.Exists() {
			str = getDataByPath(jsonPath, consts.INT, data)
			s = time.Unix(str.Int(), 0).Format(time.RFC3339)
		}

		model[path] = s
	case consts.BYTES:
		jsonPath := getGJSONPath(path, binPath, indices...)
		str := getDataByPath(jsonPath, consts.BYTES, data)
		model[path] = str.String()
	case consts.MAP, consts.BIGMAP:
		jsonPath := getGJSONPath(path, binPath, indices...)
		mapData := data.Get(jsonPath).Array()
		result := make([]interface{}, 0)
		for i := range mapData {
			itemModel := make(DefaultModel)
			newIndices := append(indices, i)
			for _, suffix := range []string{"/k", "/v"} {
				keyPath := path + suffix
				keyNode, ok := metadata[keyPath]
				if !ok {
					return fmt.Errorf("Unknown map node: %s", keyPath)
				}
				if err := itemModel.fill(data, metadata, keyNode, keyPath, prefix, false, newIndices...); err != nil {
					return err
				}
			}
			result = append(result, itemModel)
		}
		model[path] = result
	case consts.OR:
		orPath := path
		jsonPath := getGJSONPath(orPath, binPath, indices...)
		end := false
		for !end {
			jsonPath = strings.TrimPrefix(jsonPath, ".")
			p := getDataByPath(jsonPath, "prim|@lower", data)
			prim := p.String()
			switch prim {
			case consts.LEFT:
				orPath += "/0"
				jsonPath += ".args.0"
			case consts.RIGHT:
				orPath += "/1"
				jsonPath += ".args.1"
			default:
				end = true
			}
		}
		model[path] = DefaultModel{
			"schemaKey": orPath,
		}
	case consts.LAMBDA:
		jsonPath := getGJSONPath(path, binPath, indices...)
		str, err := formatter.MichelineToMichelson(data.Get(jsonPath), false, formatter.DefLineSize)
		if err != nil {
			return err
		}
		model[path] = str
	case consts.BOOL:
		jsonPath := getGJSONPath(path, binPath, indices...)
		b := getDataByPath(jsonPath, "prim|@lower", data)
		model[path] = b.Bool()
	case consts.OPTION:
		optionNode, ok := metadata[path]
		if !ok {
			return fmt.Errorf("Unknown option node: %s", path)
		}
		if !strings.HasSuffix(path, "/o") {
			path += "/o"
		}
		if err := model.fill(data, metadata, optionNode, path, prefix, false, indices...); err != nil {
			return err
		}
	default:
	}
	return nil
}

func getGJSONPath(fullPath, path string, indices ...int) string {
	if fullPath == "0" || path == "" {
		return ""
	}
	jsonPath := newmiguel.GetGJSONPathForData(path)
	for i := range indices {
		jsonPath = strings.Replace(jsonPath, "#", fmt.Sprintf("%d", indices[i]), 1)
	}
	return jsonPath
}

func getDataByPath(path, suffix string, data gjson.Result) gjson.Result {
	p := suffix
	if path != "" {
		p = fmt.Sprintf("%s.%s", path, p)
	}
	return data.Get(p)
}

func (model DefaultModel) optionWrapper(data gjson.Result, metadata meta.Metadata, node *meta.NodeMetadata, path, prefix string, indices ...int) error {
	if !strings.HasSuffix(path, "/o") {
		return nil
	}
	binPath := prepareBinPath(path, prefix)
	jsonPath := getGJSONPath(path, binPath, indices...)
	if !data.Get(jsonPath).Exists() {
		model[strings.TrimSuffix(path, "/o")] = DefaultModel{
			"schemaKey": consts.NONE,
		}
		return nil
	}

	optionModel := DefaultModel{
		"schemaKey": consts.SOME,
	}
	if err := optionModel.fill(data, metadata, node, path, prefix, true, indices...); err != nil {
		return err
	}
	model[strings.TrimSuffix(path, "/o")] = optionModel
	return nil
}

func prepareBinPath(path, prefix string) string {
	binPath := strings.TrimPrefix(path, prefix)
	if path == "0" {
		binPath = path
	}

	if strings.HasPrefix(binPath, "/") {
		binPath = strings.TrimPrefix(binPath, "/")
	}
	return binPath
}
