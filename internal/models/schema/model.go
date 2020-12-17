package schema

// Schema -
type Schema struct {
	ID        string            `json:"-"`
	Parameter map[string]string `json:"parameter"`
	Storage   map[string]string `json:"storage"`
}

// GetID -
func (m *Schema) GetID() string {
	return m.ID
}

// GetIndex -
func (m *Schema) GetIndex() string {
	return "schema"
}

// GetQueues -
func (m *Schema) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (m *Schema) MarshalToQueue() ([]byte, error) {
	return nil, nil
}
