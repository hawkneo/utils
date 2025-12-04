package migrate

import (
	"context"
	"database/sql"
	"github.com/gridexswap/utils/log"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDirectoryMigrationSource_LoadMigrations(t *testing.T) {
	source := DirectoryMigrationSource{Directory: "./test_data"}
	migrations, err := source.LoadMigrations()
	require.NoError(t, err)
	require.True(t, len(migrations) > 0)
}

func TestStatusMigrator(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=gridex sslmode=disable TimeZone=Etc/UTC")
	require.NoError(t, err)
	ctx := &Context{
		Context: context.TODO(),
		Conf: &Config{
			Dialect:           PostgresDialect{},
			DB:                db,
			SchemaName:        "migration_schema",
			Logger:            log.AnsiColorLogger{ColorOutput: true},
			MigrateOutOfOrder: false,
			MigrationSource:   DirectoryMigrationSource{Directory: "./test_data"},
		},
	}
	migrator, err := NewStatusMigrator(ctx)
	require.NoError(t, err)
	require.NoError(t, migrator.Apply())
}

func TestUpMigrator(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=gridex sslmode=disable TimeZone=Etc/UTC")
	require.NoError(t, err)
	ctx := &Context{
		Context: context.TODO(),
		Conf: &Config{
			Dialect:           PostgresDialect{},
			DB:                db,
			SchemaName:        "migration_schema",
			MigrateOutOfOrder: false,
			MigrationSource:   DirectoryMigrationSource{Directory: "./test_data"},
			DryRun:            true,
		},
	}

	migrator, err := NewUpMigrator(ctx)
	require.NoError(t, err)
	require.NoError(t, migrator.Apply())
}

func TestDownMigrator(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=gridex sslmode=disable TimeZone=Etc/UTC")
	require.NoError(t, err)
	ctx := &Context{
		Context: context.TODO(),
		Conf: &Config{
			Dialect:           PostgresDialect{},
			DB:                db,
			SchemaName:        "migration_schema",
			MigrateOutOfOrder: false,
			MigrationSource:   DirectoryMigrationSource{Directory: "./test_data"},
		},
	}

	migrator, err := NewDownMigrator(ctx, "20221125191249")
	require.NoError(t, err)
	require.NoError(t, migrator.Apply())
}
