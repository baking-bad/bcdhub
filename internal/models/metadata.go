package models

// Metadata -
type Metadata struct {
	ID        string            `json:"-"`
	Parameter map[string]string `json:"parameter"`
	Storage   map[string]string `json:"storage"`
}

// GetID -
func (m *Metadata) GetID() string {
	return m.ID
}

// GetIndex -
func (m *Metadata) GetIndex() string {
	return "metadata"
}

// GetQueues -
func (m *Metadata) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (m *Metadata) MarshalToQueue() ([]byte, error) {
	return nil, nil
}
