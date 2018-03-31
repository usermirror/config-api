package config

// CampaignConfig ...
type CampaignConfig struct {
	NamespaceID string      `json:"namespace_id"`
	ConfigID    string      `json:"config_id"`
	Type        string      `json:"type,omitempty"`
	Body        interface{} `json:"body,omitempty"`
}
