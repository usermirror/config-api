package postgres

import (
	"github.com/usermirror/config-api/pkg/migrations/schema"
)

var createNamespaceTokens = schema.Migration{
	Name:    "create_namespace_tokens",
	Version: 2,
	Up: func() string {
		return `CREATE TABLE IF NOT EXISTS "public"."namespace_tokens" (
	"namespace_id" bytea,
	"write_tokens" bytea,
	PRIMARY KEY (namespace_id)
);`
	},
	Down: func() string {
		return `DROP TABLE "public"."namespace_tokens";`
	},
}
