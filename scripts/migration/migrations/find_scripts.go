package migrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd"
	astContract "github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// FindLostContracts -
type FindLostContracts struct {
	lastID    int64
	bulkCount int
}

// Key -
func (m *FindLostContracts) Key() string {
	return "find_lost_script_relations"
}

// Description -
func (m *FindLostContracts) Description() string {
	return "find `scripts` relation"
}

// Do - migrate function
func (m *FindLostContracts) Do(ctx *config.Context) error {
	if m.bulkCount == 0 {
		m.bulkCount = 1000
	}

	return ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		var end bool

		for !end {
			fmt.Printf("last id = %d\r", m.lastID)
			contracts, err := m.getContracts(tx)
			if err != nil {
				return err
			}

			if err := m.findScripts(ctx, tx, contracts); err != nil {
				return err
			}

			end = len(contracts) != m.bulkCount

			if err := m.saveContracts(tx, contracts); err != nil {
				return err
			}
		}
		return nil
	})

}

func (m *FindLostContracts) getContracts(db pg.DBI) (resp []contract.Contract, err error) {
	query := db.Model((*contract.Contract)(nil)).Order("contract.id asc").Where("alpha_id is NULL and babylon_id is null").Relation("Account.address")
	if m.lastID > 0 {
		query.Where("contract.id > ?", m.lastID)
	}
	err = query.Limit(m.bulkCount).Select(&resp)
	return
}

func (m *FindLostContracts) saveContracts(db pg.DBI, contracts []contract.Contract) (err error) {
	_, err = db.Model(&contracts).Set("alpha_id = ?alpha_id, babylon_id = ?babylon_id, tags = ?tags").WherePK().Update()
	return
}

func (m *FindLostContracts) findScripts(ctx *config.Context, db pg.DBI, contracts []contract.Contract) error {
	for i := range contracts {
		rpc, err := ctx.GetRPC(contracts[i].Network)
		if err != nil {
			return err
		}
		scriptBytes, err := rpc.GetRawScript(contracts[i].Account.Address, 0)
		if err != nil {
			return err
		}
		script, err := astContract.NewParser(scriptBytes)
		if err != nil {
			return errors.Wrap(err, "astContract.NewParser")
		}
		contractScript, err := ctx.Scripts.ByHash(script.Hash)
		if err != nil {
			if !ctx.Storage.IsRecordNotFound(err) {
				return err
			}
			var s bcd.RawScript
			if err := json.Unmarshal(script.CodeRaw, &s); err != nil {
				return err
			}
			contractScript = contract.Script{
				Hash:      script.Hash,
				Code:      s.Code,
				Parameter: s.Parameter,
				Storage:   s.Storage,
				Views:     s.Views,
			}

			constants, err := script.FindConstants()
			if err != nil {
				return errors.Wrap(err, "script.FindConstants")
			}

			if len(constants) > 0 {
				globalConstants, err := ctx.GlobalConstants.All(contracts[i].Network, constants...)
				if err != nil {
					return err
				}
				contractScript.Constants = globalConstants
				scriptBytes = m.replaceConstants(&contractScript, scriptBytes)

				script, err = astContract.NewParser(scriptBytes)
				if err != nil {
					return errors.Wrap(err, "astContract.NewParser")
				}

				if err := script.Parse(); err != nil {
					return err
				}

				contractScript.FingerprintParameter = script.Fingerprint.Parameter
				contractScript.FingerprintCode = script.Fingerprint.Code
				contractScript.FingerprintStorage = script.Fingerprint.Storage
				contractScript.FailStrings = script.FailStrings.Values()
				contractScript.Annotations = script.Annotations.Values()
				contractScript.Tags = types.NewTags(script.Tags.Values())
				contractScript.Hardcoded = script.HardcodedAddresses.Values()

				params, err := script.Code.Parameter.ToTypedAST()
				if err != nil {
					return err
				}
				contractScript.Entrypoints = params.GetEntrypoints()

				if script.IsUpgradable() {
					contractScript.Tags.Set(types.UpgradableTag)
				}

				if err := contractScript.Save(db); err != nil {
					return err
				}

				if contracts[i].Network != types.Mainnet {
					contracts[i].BabylonID = contractScript.ID
				} else {
					if contracts[i].Level <= 655360 {
						contracts[i].AlphaID = contractScript.ID
					} else {
						contracts[i].BabylonID = contractScript.ID
					}
				}

			}
			continue
		}
		if contracts[i].Network != types.Mainnet {
			contracts[i].BabylonID = contractScript.ID
		} else {
			if contracts[i].Level <= 655360 {
				contracts[i].AlphaID = contractScript.ID
			} else {
				contracts[i].BabylonID = contractScript.ID
			}
		}
	}

	return nil
}

func (m *FindLostContracts) replaceConstants(c *contract.Script, script []byte) []byte {
	pattern := `{"prim":"constant","args":[{"string":"%s"}]}`
	for i := range c.Constants {
		script = bytes.ReplaceAll(
			script,
			[]byte(fmt.Sprintf(pattern, c.Constants[i].Address)),
			c.Constants[i].Value,
		)
	}
	return script
}
