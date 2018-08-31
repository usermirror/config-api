package storage

// Store allows persistence of arbitrary data.
type Store interface {
	Init() error
	Get(input GetInput) ([]byte, error)
	Set(input SetInput) error
	Scan(input ScanInput) (KeyList, error)
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

// ScanInput is a request to scan for keys given a prefix.
type ScanInput struct {
	Prefix  string
	Timeout int
}

// KeyList is a list of key/value pairs and a cursor.
type KeyList struct {
	Keys   []string `json:"keys"`
	Cursor int
}

// Config contains options specific to an individual configuration.
type Config struct {
	NamespaceID string      `json:"namespace_id"`
	ConfigID    string      `json:"config_id"`
	Type        string      `json:"type,omitempty"`
	Body        interface{} `json:"body,omitempty"`
}
