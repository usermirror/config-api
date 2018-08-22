package config

import (
	"github.com/usermirror/config-api/pkg/storage"
)

// Get retrieves configuration data from a config-api.
func (c *Client) Get(GetInput) storage.Config {
	return storage.Config{}
}

// Set persists configuration data to a config-api.
func (c *Client) Set(SetInput) error {
	return nil
}
