package main

import (
	"strings"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/google/uuid"
)

func createRPCs(cfg config) map[string]*noderpc.NodeRPC {
	rpc := make(map[string]*noderpc.NodeRPC)
	for i := range cfg.NodeRPC {
		nodeCfg := cfg.NodeRPC[i]
		rpc[nodeCfg.Network] = noderpc.NewNodeRPC(nodeCfg.Host)
		rpc[nodeCfg.Network].SetTimeout(time.Second * 30)
	}
	return rpc
}

func updateState(es *elastic.Elastic, last models.Contract) error {
	if currentState.Timestamp.After(last.Timestamp) {
		return nil
	}
	currentState.Timestamp = last.Timestamp

	if _, err := es.UpdateDoc(elastic.DocStates, currentState.ID, currentState); err != nil {
		return err
	}
	return nil
}

func getContractProjectID(es *elastic.Elastic, c models.Contract, buckets []models.Contract) (string, error) {
	for i := len(buckets) - 1; i > -1; i-- {
		ok, err := compare(c, buckets[i])
		if err != nil {
			return "", err
		}

		if ok {
			return buckets[i].ProjectID, nil
		}
	}

	projID := strings.ReplaceAll(uuid.New().String(), "-", "")
	proj := models.Project{
		ID:    projID,
		Alias: projID,
	}

	if _, err := es.AddDocumentWithID(proj, elastic.DocProjects, projID); err != nil {
		return "", err
	}

	return projID, nil
}

func sync(rpcs map[string]*noderpc.NodeRPC, es *elastic.Elastic) error {
	logger.Info("Current state: %s", currentState.Timestamp.String())

	contracts, err := es.GetContractsByTime(currentState.Timestamp, "asc")
	if err != nil {
		return err
	}

	logger.Info("Found %d contracts", len(contracts))

	var buckets []models.Contract
	if !currentState.Timestamp.IsZero() {
		buckets, err = es.GetLastProjectContracts()
		if err != nil {
			return err
		}
	} else {
		buckets = make([]models.Contract, 0)
	}

	for _, c := range contracts {
		fgpt, err := computeFingerprint(rpcs[c.Network], c)
		if err != nil {
			return err
		}
		c.Fingerprint = &fgpt

		projID, err := getContractProjectID(es, c, buckets)
		if err != nil {
			return err
		}
		c.ProjectID = projID
		buckets = append(buckets, c)

		logger.Info("Contract %s to project %s", c.Address, c.ProjectID)

		if _, err := es.UpdateDoc(elastic.DocContracts, c.ID, c); err != nil {
			return err
		}

		if err := updateState(es, c); err != nil {
			return err
		}
	}

	logger.Success("Synced")
	return nil
}
