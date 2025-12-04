package migrate

import (
	"database/sql"
	"github.com/gridexswap/utils/log"
)

type Config struct {
	Dialect Dialect
	DB      *sql.DB
	// Logger is used to log messages.
	Logger log.Logger
	// SchemaName is the name of the schema that will be used to store the migration history.
	SchemaName  string
	schemaMaxID uint64

	// Indicates whether to print status
	printStatus bool

	// MigrateOutOfOrder is a flag that if you already have versions 1.0 and 3.0 applied, and now a version 2.0 is found, it will be applied too instead of being ignored.
	MigrateOutOfOrder bool

	MigrationSource MigrationSource

	// DryRun is a flag that if set to true, will not apply any migrations, but will instead print out what migrations would have been applied.
	DryRun bool
}
