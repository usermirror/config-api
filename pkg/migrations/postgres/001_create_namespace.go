package postgres

import (
	"github.com/usermirror/config-api/pkg/migrations/schema"
)

var createNamespaces = schema.Migration{
	Name:    "create_namespaces",
	Version: 2,
	Up: func() string {
		return `CREATE TABLE IF NOT EXISTS "public"."namespaces" (
	"namespace_id" bytea,
	"write_tokens" bytea,
	PRIMARY KEY (namespace_id)
);`
	},
	Down: func() string {
		return `DROP TABLE "public"."namespaces";`
	},
}
