package migrations

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"

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

// SetFA1 - migration that set fa1 tag to contract
type SetFA1 struct{}

// Description -
func (m *SetFA1) Description() string {
	return "set fa1 tag to contract"
}

// Do - migrate function
func (m *SetFA1) Do(ctx *Context) error {
	contracts, err := ctx.ES.GetContracts(nil)
	if err != nil {
		return err
	}

	logger.Info("Loading FA1 interface...")
	var fa1 []contractparser.Entrypoint
	if err := json.Unmarshal(fa1Interface, &fa1); err != nil {
		return err
	}

	logger.Info("Found %d contracts", len(contracts))

	bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false))
	for _, c := range contracts {
		bar.Add(1)
		m, err := meta.GetMetadata(ctx.ES, c.Address, consts.PARAMETER, "PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS")
		if err != nil {
			if !strings.Contains(err.Error(), "Unknown metadata sym link") {
				fmt.Print("\033[2K\r")
				return err
			}
			m, err = meta.GetMetadata(ctx.ES, c.Address, consts.PARAMETER, "PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex")
			if err != nil {
				fmt.Print("\033[2K\r")
				return err
			}
		}
		if !findInterface(m, fa1) {
			continue
		}

		c.Tags = append(c.Tags, consts.FA1Tag)

		if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, c.ID, c); err != nil {
			fmt.Print("\033[2K\r")
			logger.Errorf("ctx.ES.UpdateDoc %v %v error: %v", c.ID, c, err)
			return err
		}
	}

	fmt.Print("\033[2K\r")
	logger.Info("Done. Total contracts: %d", len(contracts))

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
