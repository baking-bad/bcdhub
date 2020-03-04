package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/aopoltorzhicky/bcdhub/internal/classification/metrics"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
)

var model = []metrics.Metric{
	metrics.NewBool("Manager"),
	metrics.NewArray("Tags"),
	metrics.NewArray("FailStrings"),
	metrics.NewArray("Annotations"),
	metrics.NewBool("Language"),
	metrics.NewArray("Entrypoints"),
	metrics.NewFingerprintLength("parameter"),
	metrics.NewFingerprintLength("storage"),
	metrics.NewFingerprintLength("code"),
	metrics.NewFingerprint("parameter"),
	metrics.NewFingerprint("storage"),
	metrics.NewFingerprint("code"),
}

const (
	fileName = "train.csv"
)

var mux sync.Mutex

// Vote -
func (ctx *Context) Vote(c *gin.Context) {
	var req voteRequest
	if err := c.BindJSON(&req); handleError(c, err, http.StatusBadRequest) {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	a, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.SourceAddress,
		"network": req.SourceNetwork,
	})
	if handleError(c, err, 0) {
		return
	}

	b, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.DestinationAddress,
		"network": req.DestinationNetwork,
	})
	if handleError(c, err, 0) {
		return
	}

	if err := compare(a, b, req.Vote); handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, "")
}

func compare(a, b models.Contract, vote int) error {
	mux.Lock()
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	sum := 0.0
	features := make([]float64, len(model))

	record := make([]string, len(model)+1)
	for i := range model {
		f := model[i].Compute(a, b)
		features[i] = f.Value
		record[i] = fmt.Sprintf("%v", f.Value)

		if sum > 1 {
			return fmt.Errorf("Invalid metric weights. Check sum of weight is not equal 1")
		}
	}
	record[len(record)-1] = fmt.Sprintf("%d", vote)

	if err := w.Write(record); err != nil {
		return err
	}
	w.Flush()
	mux.Unlock()
	return nil
}
