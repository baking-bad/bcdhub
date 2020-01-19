package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/google/uuid"
)

func computeMetrics(rpc *noderpc.NodeRPC, es *elastic.Elastic, c *models.Contract) error {
	contract, err := rpc.GetScriptJSON(c.Address, 0)
	if err != nil {
		return err
	}

	script, err := contractparser.New(contract)
	if err != nil {
		return fmt.Errorf("contractparser.New: %v", err)
	}
	script.Parse()

	c.Language = script.Language()
	c.Hash = []string{
		script.Code.Parameter.Hash,
		script.Code.Hash,
		script.Code.Storage.Hash,
	}
	c.FailStrings = script.Code.FailStrings.Values()
	c.Primitives = script.Code.Primitives.Values()
	c.Annotations = script.Code.Annotations.Values()
	c.Entrypoints = script.Code.Parameter.Entrypoints()
	c.Tags = script.Tags.Values()

	c.Hardcoded = script.HardcodedAddresses.Values()

	if err := saveMetadatas(es, rpc, c); err != nil {
		return err
	}

	err = findProject(es, c)
	return err
}

func findProject(es *elastic.Elastic, c *models.Contract) error {
	contracts, err := es.FindProjectContracts(c.Hash, 5)
	if err != nil {
		return err
	}

	var meta []byte
	bulk := bytes.NewBuffer([]byte{})

	if len(contracts) > 0 {
		c.ProjectID = contracts[0].ProjectID
		project, err := es.GetProject(c.ProjectID)
		if err != nil {
			return err
		}
		project.Contracts = append(project.Contracts, c.ID)
		if _, err = es.UpdateDoc(elastic.DocProjects, project.ID, project); err != nil {
			return err
		}
		return nil
	}
	project := models.Project{
		ID:        uuid.New().String(),
		Contracts: make([]string, 0),
	}
	c.ProjectID = project.ID
	meta = []byte(fmt.Sprintf(`{ "index" : { "_id": "%s"} }%s`, project.ID, "\n"))
	project.Contracts = append(project.Contracts, c.ID)

	data, err := json.Marshal(project)
	if err != nil {
		return err
	}
	data = append(data, "\n"...)

	bulk.Grow(len(meta) + len(data))
	bulk.Write(meta)
	bulk.Write(data)

	if err := es.BulkInsert(elastic.DocProjects, bulk); err != nil {
		return err
	}
	return nil
}
