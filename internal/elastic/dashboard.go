package elastic

import (
	"fmt"
)

// GetTimeline -
func (e *Elastic) GetTimeline(projectIDs []string, contracts []string) ([]TimelineItem, error) {
	projectContracts, err := e.getProjectsContracts(projectIDs)
	if err != nil {
		return nil, nil
	}

	contractIDs, err := e.getContractIDs(contracts)
	if err != nil {
		return nil, err
	}
	data := mergeContractIDs(projectContracts, contractIDs)

	query := "SELECT network, hash, status, timestamp, kind, source, destination, entrypoint, amount FROM operation WHERE "
	for i := range data {
		query += fmt.Sprintf("(network = '%s' AND (destination = '%s' OR source = '%s'))", data[i].Network, data[i].Address, data[i].Address)
		if i != len(data)-1 {
			query += " OR "
		}
	}
	query += " ORDER BY timestamp DESC"
	result, err := e.executeSQL(query)
	if err != nil {
		return nil, err
	}

	ops := make([]TimelineItem, 0)
	for _, hit := range result.Get("rows").Array() {
		var ti TimelineItem
		ti.ParseElasticJSONArray(hit)
		ops = append(ops, ti)
	}
	return ops, nil
}

func mergeContractIDs(a, b []ContractID) []ContractID {
	result := b
	for i := range a {
		found := false
		for j := range b {
			if b[j].Address == a[i].Address && b[j].Network == a[i].Network {
				found = true
				break
			}
		}
		if !found {
			result = append(result, a[i])
		}
	}
	return result
}

func (e *Elastic) getContractIDs(contracts []string) ([]ContractID, error) {
	query := "SELECT address, network FROM contract WHERE id IN ("
	for i := range contracts {
		query += fmt.Sprintf("'%s'", contracts[i])
		if i != len(contracts)-1 {
			query += ","
		}
	}
	query += ")"
	result, err := e.executeSQL(query)
	if err != nil {
		return nil, err
	}

	contractIDs := make([]ContractID, 0)
	for _, hit := range result.Get("rows").Array() {
		var cid ContractID
		cid.ParseElasticJSONArray(hit)
		contractIDs = append(contractIDs, cid)
	}
	return contractIDs, nil
}
