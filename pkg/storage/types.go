package storage

// Store allows persistence of arbitrary data.
type Store interface {
	Get(input GetInput) ([]byte, error)
	Set(input SetInput) error
}

// GetInput is a request to read data from a Store.
type GetInput struct {
	Key     string
	Timeout int
}

// SetInput is a request to write data to a Store.
type SetInput struct {
	Key     string
	Value   []byte
	Timeout int
}

// CampaignConfig contains options specific to an individual campaign.
type CampaignConfig struct {
	NamespaceID string      `json:"namespace_id"`
	ConfigID    string      `json:"config_id"`
	Type        string      `json:"type,omitempty"`
	Body        interface{} `json:"body,omitempty"`
}
