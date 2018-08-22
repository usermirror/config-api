package config

// Client is the Go interface for using a config-api.
type Client struct {
	// ApiHost is the address of the config-api server
	APIHost string
}

// GetInput is a request to read data from a config-api.
type GetInput struct {
	NamespaceID string
	ConfigID    string
	Timeout     int
}

// SetInput is a request to write data to a config-api.
type SetInput struct {
	NamespaceID string
	ConfigID    string
	Value       interface{}
	Timeout     int
}
