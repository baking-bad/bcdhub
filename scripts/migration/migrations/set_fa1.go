package migrations

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
)

var fa1Interface = []byte(`[
    {
        "name": "getBalance",
        "prim": "pair",
        "args": [
            {
                "prim": "address"
            },
            {
                "parameter": {
                    "prim": "nat"
                },
                "prim": "contract"
            }
        ]
    },
    {
        "name": "getTotalSupply",
        "prim": "pair",
        "args": [
            {
                "prim": "unit"
            },
            {
                "parameter": {
                    "prim": "nat"
                },
                "prim": "contract"
            }
        ]
    },
    {
        "name": "transfer",
        "prim": "pair",
        "args": [
            {
                "prim": "address"
            },
            {
                "args": [
                    {
                        "prim": "address"
                    },
                    {
                        "prim": "nat"
                    }
                ],
                "prim": "pair"
            }
        ]
    }
]`)

// SetFA1Migration - migration that set fa1 tag to contract
type SetFA1Migration struct{}

// Do - migrate function
func (m *SetFA1Migration) Do(ctx *Context) error {
	log.Print("Start SetFA1Migration...")
	contracts, err := ctx.ES.GetContracts(nil)
	if err != nil {
		return err
	}

	log.Print("Loading FA1 interface...")
	var fa1 []contractparser.Entrypoint
	if err := json.Unmarshal(fa1Interface, &fa1); err != nil {
		return err
	}

	log.Printf("Found %d contracts", len(contracts))
	for i, c := range contracts {
		m, err := meta.GetMetadata(ctx.ES, c.Address, consts.PARAMETER, consts.HashBabylon)
		if err != nil {
			if !strings.Contains(err.Error(), "Unknown metadata sym link") {
				return err
			}
			m, err = meta.GetMetadata(ctx.ES, c.Address, consts.PARAMETER, consts.Hash1)
			if err != nil {
				return err
			}
		}
		if !findInterface(m, fa1) {
			continue
		}

		c.Tags = append(c.Tags, consts.FA1Tag)

		if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, c.ID, c); err != nil {
			log.Println("ctx.ES.UpdateDoc error:", c.ID, c, err)
			return err
		}
		log.Printf("%d/%d | %v", i, len(contracts), c.ID)
	}

	return nil
}

func findInterface(metadata meta.Metadata, i []contractparser.Entrypoint) bool {
	root := metadata["0"]

	for _, ie := range i {
		found := false
		for _, e := range root.Args {
			entrypointMeta := metadata[e]
			if compareEntrypoints(metadata, ie, *entrypointMeta, e) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func compareEntrypoints(metadata meta.Metadata, in contractparser.Entrypoint, en meta.NodeMetadata, path string) bool {
	if in.Name != "" && en.Name != in.Name {
		return false
	}
	// fmt.Printf("[in] %+v\n[en] %+v\n\n", in, en)
	if in.Prim != en.Prim {
		return false
	}

	for i, inArg := range in.Args {
		enPath := fmt.Sprintf("%s/%d", path, i)
		enMeta, ok := metadata[enPath]
		if !ok {
			return false
		}
		if !compareEntrypoints(metadata, inArg, *enMeta, enPath) {
			return false
		}
	}

	return true
}
