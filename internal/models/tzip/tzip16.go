package tzip

import (
	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/tidwall/gjson"
)

// TZIP16 -
type TZIP16 struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Version     string      `json:"version,omitempty"`
	License     License     `json:"license,omitempty"`
	Homepage    string      `json:"homepage,omitempty"`
	Authors     []string    `json:"authors,omitempty"`
	Interfaces  []string    `json:"interfaces,omitempty"`
	Views       interface{} `json:"views,omitempty"`
}

// ParseElasticJSON -
func (t *TZIP16) ParseElasticJSON(resp gjson.Result) {
	t.Name = resp.Get("_source.name").String()
	t.Description = resp.Get("_source.description").String()
	t.Version = resp.Get("_source.version").String()
	t.Homepage = resp.Get("_source.homepage").String()

	t.Authors = utils.StringArray(resp, "_source.authors")
	t.Interfaces = utils.StringArray(resp, "_source.interfaces")
	t.Views = resp.Get("_source.views").Value()

	t.License.ParseElasticJSON(resp)
}

// License -
type License struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

// ParseElasticJSON -
func (license *License) ParseElasticJSON(resp gjson.Result) {
	license.Name = resp.Get("name").String()
	license.Details = resp.Get("details").String()
}
