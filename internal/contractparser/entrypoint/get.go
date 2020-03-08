package entrypoint

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

// Get - returns entrypoint name by parameters node
func Get(node gjson.Result, metadata meta.Metadata) (string, error) {
	var entrypoint string
	if node.Get("entrypoint").Exists() {
		entrypoint = node.Get("entrypoint").String()
		node = node.Get("value")
	}

	startPath := "0"
	if entrypoint != "" {
		for key, nm := range metadata {
			if nm.FieldName == entrypoint {
				startPath = key
				break
			}
		}
	}
	path := getPath(node, startPath)
	eMeta, ok := metadata[path]
	if !ok {
		return entrypoint, fmt.Errorf("Invalid parameter: %s", node.String())
	}
	if eMeta.Name != "" {
		return eMeta.Name, nil
	}

	if entrypoint == "" {
		if path == "0" {
			return "default", nil
		}

		root := metadata["0"]
		for i := range root.Args {
			if root.Args[i] == path {
				return eMeta.GetEntrypointName(i), nil
			}
		}
	}

	return entrypoint, nil
}

func getPath(node gjson.Result, path string) string {
	prim := node.Get("prim").String()
	if prim == "Left" {
		path += "/0"
		subNode := node.Get("args.0")
		return getPath(subNode, path)
	}

	if prim == "Right" {
		path += "/1"
		subNode := node.Get("args.0")
		return getPath(subNode, path)
	}

	if prim == "None" || prim == "Some" {
		return path + "/o"
	}
	return path
}
