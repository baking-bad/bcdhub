package elastic

import "github.com/baking-bad/bcdhub/internal/models"

// GetMetadata -
func (e *Elastic) GetMetadata(address string) (metadata models.Metadata, err error) {
	data, err := e.GetByID(DocMetadata, address)
	if err != nil {
		return
	}
	metadata.ParseElasticJSON(data)
	return
}

// GetAllMetadata -
func (e *Elastic) GetAllMetadata(q map[string]interface{}) ([]models.Metadata, error) {
	metadata := make([]models.Metadata, 0)

	result, err := e.createScroll(DocMetadata, 1000, q)
	if err != nil {
		return nil, err
	}
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, item := range hits.Array() {
			var m models.Metadata
			m.ParseElasticJSON(item)
			metadata = append(metadata, m)
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
	}

	return metadata, nil
}
