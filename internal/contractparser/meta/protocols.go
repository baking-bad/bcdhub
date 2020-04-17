package meta

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

var symLinks = map[string]string{
	"PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i": "alpha",
	"PtBMwNZT94N7gXKw4i273CKcSaBrrBnqnt3RATExNKr9KNX2USV": "alpha",
	"ProtoDemoNoopsDemoNoopsDemoNoopsDemoNoopsDemo6XBoYp": "alpha",
	"PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex": "alpha",
	"Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P": "alpha",
	"PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt": "alpha",
	"PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP": "alpha",
	"Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd": "alpha",
	"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY": "alpha",
	"PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS": "babylon",
	"PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU": "babylon",
	"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb": "babylon",
}

var protocols map[string]string

// LoadProtocols -
func LoadProtocols(e *elastic.Elastic) error {
	response, err := e.GetProtocols()
	if err != nil {
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}
	if len(response) == 0 {
		response, err = createProtocols(e)
		if err != nil {
			return err
		}
	}
	protocols = make(map[string]string)
	for _, p := range response {
		protocols[p.Hash] = p.SymLink
	}
	return nil
}

func createProtocols(es *elastic.Elastic) ([]models.Protocol, error) {
	protocols := make([]models.Protocol, 0)
	for _, network := range []string{"mainnet", "zeronet", "carthagenet", "babylonnet"} {
		api := tzkt.NewTzKTForNetwork(network, time.Minute)

		result, err := api.GetProtocols()
		if err != nil {
			return nil, err
		}

		for i := range result {
			symLink, ok := symLinks[result[i].Hash]
			if !ok {
				return nil, fmt.Errorf("Unknown protocol: %s", result[i].Hash)
			}
			protocols = append(protocols, models.Protocol{
				ID:         helpers.GenerateID(),
				Hash:       result[i].Hash,
				Alias:      result[i].Metadata.Alias,
				StartLevel: result[i].StartLevel,
				EndLevel:   result[i].LastLevel,
				SymLink:    symLink,
				Network:    network,
			})
		}

	}
	if err := es.BulkInsertProtocols(protocols); err != nil {
		return nil, err
	}
	return protocols, nil
}
