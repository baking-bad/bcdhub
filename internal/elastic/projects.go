package elastic

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func parseProjectFormHit(hit gjson.Result, proj *models.Project) {
	proj.ID = hit.Get("_id").String()
	proj.Alias = hit.Get("_source.alias").String()
	proj.Contracts = parseStringArray(hit, "_source.contracts")
}

// GetProject -
func (e *Elastic) GetProject(id string) (p models.Project, err error) {
	res, err := e.GetByID(DocProjects, id)
	if err != nil {
		return
	}
	if !res.Get("found").Bool() {
		return p, fmt.Errorf("Unknown project: %s", id)
	}
	parseProjectFormHit(*res, &p)
	return
}
