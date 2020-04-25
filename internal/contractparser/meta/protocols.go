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

// This is the list of protocols BCD supports
// Every time new protocol is proposed we determine if everything works fine or implement a custom handler otherwise
// After that we append protocol to this list with a corresponding handler id (aka symlink)
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

// GetProtoSymLink -
func GetProtoSymLink(protocol string) (string, error) {
	if protoSymLink, ok := symLinks[protocol]; ok {
		return protoSymLink, nil
	}
	return "", fmt.Errorf("Unknown protocol: %s", protocol)
}

// LoadProtocols -
func LoadProtocols(e *elastic.Elastic, networks []string) error {
	response, err := e.GetProtocols()
	if err != nil {
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}
	if len(response) == 0 {
		response, err = createProtocols(e, networks)
		if err != nil {
			return err
		}
	}
	return nil
}

func createProtocols(es *elastic.Elastic, networks []string) ([]models.Protocol, error) {
	protocols := make([]models.Protocol, 0)
	for _, network := range networks {
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
