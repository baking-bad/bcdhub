package elastic

import (
	"fmt"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func parseStringArray(hit gjson.Result, tag string) []string {
	res := make([]string, 0)
	for _, t := range hit.Get(tag).Array() {
		res = append(res, t.String())
	}
	return res
}

func parseContractFromHit(hit gjson.Result, c *models.Contract) {
	c.ID = hit.Get("_id").String()
	c.Network = hit.Get("_source.network").String()
	c.Level = hit.Get("_source.level").Int()
	c.Timestamp = hit.Get("_source.timestamp").Time().UTC()
	c.Balance = hit.Get("_source.balance").Int()
	c.Language = hit.Get("_source.language").String()

	c.Tags = parseStringArray(hit, "_source.tags")
	c.Hardcoded = parseStringArray(hit, "_source.hardcoded")
	c.Annotations = parseStringArray(hit, "_source.annotations")
	c.Primitives = parseStringArray(hit, "_source.primitives")
	c.FailStrings = parseStringArray(hit, "_source.fail_strings")
	c.Entrypoints = parseStringArray(hit, "_source.entrypoints")
	c.Fingerprint = getFingerprint(hit.Get("_source.fingerprint"))

	c.Address = hit.Get("_source.address").String()
	c.Manager = hit.Get("_source.manager").String()
	c.Delegate = hit.Get("_source.delegate").String()

	c.ProjectID = hit.Get("_source.project_id").String()

	c.LastAction = models.BCDTime{
		Time: hit.Get("_source.last_action").Time().UTC(),
	}

	c.TxCount = hit.Get("_source.tx_count").Int()
	c.SumTxAmount = hit.Get("_source.sum_tx_amount").Int()

	c.FoundBy = getFoundBy(hit)
}

func getFingerprint(hit gjson.Result) *models.Fingerprint {
	if !hit.Exists() {
		return nil
	}

	return &models.Fingerprint{
		Code:      hit.Get("code").String(),
		Parameter: hit.Get("parameter").String(),
		Storage:   hit.Get("storage").String(),
	}
}

func getFoundBy(hit gjson.Result) string {
	keys := make([]string, 0)
	for k := range hit.Get("highlight").Map() {
		keys = append(keys, k)
	}

	if helpers.StringInArray("address", keys) {
		return "address"
	}
	if helpers.StringInArray("manager", keys) {
		return "manager"
	}
	if helpers.StringInArray("delegate", keys) {
		return "delegate"
	}
	if helpers.StringInArray("tags", keys) {
		return "tags"
	}
	if helpers.StringInArray("hardcoded", keys) {
		return "hardcoded addresses"
	}
	if helpers.StringInArray("annotations", keys) {
		return "annotations"
	}
	if helpers.StringInArray("fail_strings", keys) {
		return "fail strings"
	}
	if helpers.StringInArray("entrypoints", keys) {
		return "entrypoints"
	}
	return ""
}

func getContractQuery(by map[string]interface{}) base {
	matches := make([]qItem, 0)
	for k, v := range by {
		matches = append(matches, matchPhrase(k, v))
	}
	return newQuery().Query(
		boolQ(
			must(matches...),
		),
	)
}

func (e *Elastic) getContract(q map[string]interface{}) (c models.Contract, err error) {
	res, err := e.query(DocContracts, q)
	if err != nil {
		return
	}
	if res.Get("hits.total.value").Int() < 1 {
		return c, fmt.Errorf("Unknown contract: %v", q)
	}
	hit := res.Get("hits.hits.0")
	parseContractFromHit(hit, &c)
	return
}

func (e *Elastic) getContracts(q map[string]interface{}) ([]models.Contract, error) {
	res, err := e.query(DocContracts, q)
	if err != nil {
		return nil, err
	}

	contracts := make([]models.Contract, 0)
	arr := res.Get("hits.hits").Array()
	for i := range arr {
		var c models.Contract
		parseContractFromHit(arr[i], &c)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetContract -
func (e *Elastic) GetContract(by map[string]interface{}) (models.Contract, error) {
	query := getContractQuery(by).One()
	return e.getContract(query)
}

// GetContractByID -
func (e *Elastic) GetContractByID(id string) (c models.Contract, err error) {
	resp, err := e.GetByID(DocContracts, id)
	if err != nil {
		return
	}
	if !resp.Get("found").Bool() {
		return c, fmt.Errorf("Unknown contract with ID %s", id)
	}
	parseContractFromHit(*resp, &c)
	return
}

// GetContractsByTime -
func (e *Elastic) GetContractsByTime(ts time.Time, sort string) ([]models.Contract, error) {
	query := newQuery().
		Query(
			boolQ(
				must(
					rangeQ("timestamp", qItem{"gt": ts}),
				),
			),
		).
		Sort("timestamp", sort).All()

	return e.getContracts(query)
}

// GetContractField -
func (e *Elastic) GetContractField(by map[string]interface{}, field string) (interface{}, error) {
	query := getContractQuery(by).One()
	res, err := e.query(DocContracts, query, field)
	if err != nil {
		return nil, err
	}
	if res.Get("hits.total.value").Int() < 1 {
		return nil, fmt.Errorf("Unknown contract: %v", by)
	}
	return res.Get("hits.hits.0._source").Get(field).Value(), nil
}

func parseContracts(res *gjson.Result) []models.Contract {
	contracts := make([]models.Contract, 0)
	arr := res.Get("hits.hits").Array()
	for i := range arr {
		var c models.Contract
		parseContractFromHit(arr[i], &c)
		contracts = append(contracts, c)
	}
	return contracts
}

// GetContracts -
func (e *Elastic) GetContracts(by map[string]interface{}) ([]models.Contract, error) {
	query := getContractQuery(by).All()
	return e.getContracts(query)
}

// GetRandomContract -
func (e *Elastic) GetRandomContract() (models.Contract, error) {
	query := newQuery().Query(qItem{
		"function_score": qItem{
			"functions": []qItem{
				qItem{
					"random_score": qItem{
						"seed": time.Now().UnixNano(),
					},
				},
			},
		},
	}).One()
	return e.getContract(query)
}

// GetContractStats -
func (e *Elastic) GetContractStats(address, network string) (stats ContractStats, err error) {
	b := boolQ(
		must(
			matchPhrase("network", network),
		),
		should(
			matchPhrase("source", address),
			matchPhrase("destination", address),
		),
	)
	b.Get("bool").Append("minimum_should_match", 1)
	query := newQuery().Query(b).Add(
		qItem{
			"aggs": qItem{
				"last_action":   max("timestamp"),
				"tx_count":      count("level"),
				"sum_tx_amount": sum("amount"),
			},
		},
	).Zero()
	res, err := e.query(DocOperations, query)
	if err != nil {
		return
	}
	stats.parse(res.Get("aggregations"))
	return
}
