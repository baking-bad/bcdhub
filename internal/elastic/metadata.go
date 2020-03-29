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
