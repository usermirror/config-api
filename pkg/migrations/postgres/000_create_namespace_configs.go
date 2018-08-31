package postgres

import (
	"github.com/usermirror/config-api/pkg/migrations/schema"
)

var createNamespaceConfigs = schema.Migration{
	Name:    "create_namespace_configs",
	Version: 1,
	Up: func() string {
		return `CREATE TABLE IF NOT EXISTS "public"."namespace_configs" (
	"key" bytea,
	"value" bytea,
	PRIMARY KEY (key)
);`
	},
	Down: func() string {
		return `DROP TABLE "public"."namespace_configs";`
	},
}
