package schema

// Migration defines a reflexive operation on a storage backend
type Migration struct {
	// Name of the migration
	Name string
	// Version of the current schema
	Version int
	// Migration methods
	Up   func() string
	Down func() string
}
