package postgres

import (
	"errors"

	"github.com/usermirror/config-api/pkg/migrations/schema"
)

var migrations = []schema.Migration{
	createNamespaceConfigs,
	createNamespaces,
}

// GetSQL returns the migration SQL for postgres
func GetSQL(direction string, version int) (string, error) {
	var sql string

	for _, migration := range migrations {
		// Skip the migration if it's less than the current version
		if migration.Version < version {
			continue
		}

		if direction == "up" {
			sql += "\n" + migration.Up()
		} else if direction == "down" {
			sql += "\n" + migration.Down()
		} else {
			return "", errors.New("migrate.getSQL: unknown direction used: " + direction)
		}
	}

	return sql, nil
}
