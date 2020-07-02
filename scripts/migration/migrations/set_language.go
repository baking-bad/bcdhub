package migrations

import (
	"log"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/language"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetLanguage - migration that set langugage to contracts with unknown language
type SetLanguage struct{}

// Key -
func (m *SetLanguage) Key() string {
	return "language"
}

// Description -
func (m *SetLanguage) Description() string {
	return "set langugage to contracts with unknown language"
}

// Do - migrate function
func (m *SetLanguage) Do(ctx *config.Context) error {
	filter := make(map[string]interface{})
	filter["language"] = language.LangUnknown

	contracts, err := ctx.ES.GetContracts(filter)
	if err != nil {
		return err
	}

	logger.Info("Found %d contracts", len(contracts))

	state, err := ctx.ES.GetAllStates()
	if err != nil {
		return err
	}

	logger.Info("Getting current protocols...")
	protocols := map[string]string{}
	for i := range state {
		logger.Info("%s -> %s", state[i].Network, state[i].Protocol)
		protocols[state[i].Network] = state[i].Protocol
	}

	var results []elastic.Model
	bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range contracts {
		bar.Add(1)
		rpc, err := ctx.GetRPC(contracts[i].Network)
		if err != nil {
			log.Println("ctx.GetRPC error:", contracts[i].ID, contracts[i].Network, err)
			return err
		}

		protocol := protocols[contracts[i].Network]
		rawScript, err := contractparser.GetContract(rpc, contracts[i].Address, contracts[i].Network, protocol, ctx.Config.Share.Path, 0)
		if err != nil {
			log.Println("contractparser.GetContract error:", contracts[i].ID, contracts[i].Address, err)
			return err
		}

		script, err := contractparser.New(rawScript)
		if err != nil {
			log.Println("contractparser.New error:", contracts[i].ID, contracts[i].Address, err)
			return err
		}

		lang, err := script.Language()
		if err != nil {
			log.Println("script.Language error:", contracts[i].ID, contracts[i].Address, err)
			return err
		}

		if lang == language.LangUnknown {
			continue
		}

		contracts[i].Language = lang
		results = append(results, &contracts[i])
	}

	if err := ctx.ES.BulkUpdate(results); err != nil {
		log.Println("ctx.ES.BulkUpdate error:", err)
		return err
	}

	return nil
}
