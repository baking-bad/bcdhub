package elastic

import (
	"fmt"
)

// GetTimeline -
func (e *Elastic) GetTimeline(contracts []string, size, from int64) ([]TimelineItem, error) {
	if len(contracts) == 0 {
		return []TimelineItem{}, nil
	}
	contractIDs, err := e.getContractIDs(contracts)
	if err != nil {
		return nil, err
	}

	conditions := make([]qItem, 0)
	for i := range contractIDs {
		boolItem := boolQ(
			must(matchPhrase("network", contractIDs[i].Network)),
			should(
				matchPhrase("source", contractIDs[i].Address),
				matchPhrase("destination", contractIDs[i].Address),
				matchPhrase("address", contractIDs[i].Address),
			),
			minimumShouldMatch(1),
		)
		conditions = append(conditions, boolItem)
	}

	b := boolQ(
		must(
			boolQ(
				should(
					matchPhrase("kind", "origination"),
					exists("errors"),
					matchPhrase("kind", "genesis"),
				),
				minimumShouldMatch(1),
			),
		),
		should(conditions...),
		minimumShouldMatch(1),
	)
	query := newQuery().Query(b).Size(size).From(from).Sort("timestamp", "desc")

	result, err := e.query([]string{DocOperations, DocMigrations}, query)
	if err != nil {
		return nil, err
	}

	timeline := make([]TimelineItem, 0)
	for _, hit := range result.Get("hits.hits").Array() {
		var t TimelineItem
		switch hit.Get("_index").String() {
		case DocOperations:
			t.ParseJSONOperation(hit)
		case DocMigrations:
			t.ParseJSONMigration(hit)
		default:
			return nil, fmt.Errorf("[GetTimeline] Invalid index type: %s", hit.String())
		}
		timeline = append(timeline, t)
	}
	return timeline, nil
}

func (e *Elastic) getContractIDs(contracts []string) ([]contractPair, error) {
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

	contractIDs := make([]contractPair, 0)
	for _, hit := range result.Get("rows").Array() {
		var cid contractPair
		cid.ParseElasticJSONArray(hit)
		contractIDs = append(contractIDs, cid)
	}
	return contractIDs, nil
}
