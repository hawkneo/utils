package migrate

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/gridexswap/utils/log"
	"sort"
	"strings"
	"time"
)

var (
	_ Migrator = (*createMigrator)(nil)
	_ Migrator = (*baselineMigrator)(nil)
	_ Migrator = (*upMigrator)(nil)
	_ Migrator = (*downMigrator)(nil)
	_ Migrator = (*statusMigrator)(nil)
)

type Statement string

type Status string

const (
	StatusApplied  Status = "applied"
	StatusPending  Status = "pending"
	StatusBaseline Status = "baseline"

	StatusOutOfOrder Status = "outOfOrder"

	StatusHashMismatch     Status = "hashMismatch"
	StatusFilenameMismatch Status = "filenameMismatch"
)

func (status Status) AnsiColorString() string {
	switch status {
	case StatusBaseline:
		return log.AnsiColorGreen(status).AnsiColorString()
	case StatusApplied:
		return log.AnsiColorGreen(status).AnsiColorString()
	case StatusPending:
		return log.AnsiColorBlue(status).AnsiColorString()
	case StatusOutOfOrder:
		return log.AnsiColorYellow(status).AnsiColorString()
	case StatusHashMismatch:
		return log.AnsiColorRed(status).AnsiColorString()
	case StatusFilenameMismatch:
		return log.AnsiColorRed(status).AnsiColorString()
	default:
		return log.AnsiColorRed(status).AnsiColorString()
	}
}

type Migration struct {
	Filename string
	Source   string

	version        string
	fileHash       string
	upStatements   []Statement
	downStatements []Statement
}

type Schema struct {
	ID        uint64
	Version   string
	Filename  string
	Hash      string
	Status    Status
	CreatedAt time.Time
}

type migrationStatus struct {
	migration *Migration
	schema    *Schema
	status    Status
}

type Context struct {
	context.Context
	Conf *Config

	migrations []*Migration
	ms         []migrationStatus
}

func fillContext(ctx *Context, parseUpStatement, parseDownStatement bool) error {
	if ctx.Conf.Logger == nil {
		ctx.Conf.Logger = &log.ConsoleLogger{Level: log.LevelInfo}
	}
	if ctx.Conf.SchemaName == "" {
		ctx.Conf.SchemaName = "migration_schema"
	}
	if ctx.Conf.MigrationSource == nil {
		return fmt.Errorf("migration source is not set")
	}

	migrations, err := ctx.Conf.MigrationSource.LoadMigrations()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		migration.version = SplitFilename(migration.Filename)
		migration.fileHash = fmt.Sprintf("%x", md5.Sum([]byte(migration.Source)))

		if parseUpStatement {
			migration.upStatements = splitSQLStatements(strings.NewReader(migration.Source), true)
		}
		if parseDownStatement {
			migration.downStatements = splitSQLStatements(strings.NewReader(migration.Source), false)
		}
	}
	sort.Slice(migrations, func(i, j int) bool {
		vi := migrations[i].version
		vj := migrations[j].version
		return CompareVersion(vi, vj)
	})

	// check duplicate version
	for i := 1; i < len(migrations); i++ {
		if migrations[i-1].version == migrations[i].version {
			return fmt.Errorf("duplicate version: %s", migrations[i].version)
		}
	}

	ctx.migrations = migrations
	return nil
}

type Migrator interface {
	Apply() error
}

type createMigrator struct {
	ctx *Context
}

func NewCreateMigrator(ctx *Context) (Migrator, error) {
	migrationSourceCapture := ctx.Conf.MigrationSource
	ctx.Conf.MigrationSource = StringMigrationSource{}
	defer func() {
		// restore migration source
		ctx.Conf.MigrationSource = migrationSourceCapture
	}()

	err := fillContext(ctx, false, false)
	if err != nil {
		return nil, err
	}
	return createMigrator{ctx}, nil
}

func (migrator createMigrator) Apply() error {
	db := migrator.ctx.Conf.DB
	_, err := db.ExecContext(
		migrator.ctx.Context,
		migrator.ctx.Conf.Dialect.CreateSchemaSQL(migrator.ctx.Conf.SchemaName),
	)

	migrator.ctx.Conf.Logger.Infof("%s", log.AnsiColorGreen("create schema successfully"))
	return err
}

type baselineMigrator struct {
	ctx *Context
}

func NewBaselineMigrator(ctx *Context) (Migrator, error) {
	err := fillContext(ctx, false, false)
	if err != nil {
		return nil, err
	}
	ctx.Conf.printStatus = false
	statusMigrator := &statusMigrator{ctx}
	err = statusMigrator.Apply()
	if err != nil {
		return nil, err
	}
	return baselineMigrator{ctx}, err
}

func (migrator baselineMigrator) Apply() error {
	ctx := migrator.ctx
	for _, ms := range ctx.ms {
		if ms.status != StatusPending {
			return fmt.Errorf("baseline only support pending migration files")
		}
	}
	if len(ctx.ms) == 0 {
		return fmt.Errorf("no pending migration files")
	}

	tx, err := ctx.Conf.DB.BeginTx(ctx.Context, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, ms := range ctx.ms {
		ctx.Conf.schemaMaxID++
		schema := &Schema{
			ID:        ctx.Conf.schemaMaxID,
			Version:   ms.migration.version,
			Filename:  ms.migration.Filename,
			Hash:      ms.migration.fileHash,
			Status:    StatusBaseline,
			CreatedAt: time.Now(),
		}

		_, err = tx.ExecContext(
			ctx.Context,
			ctx.Conf.Dialect.InsertSchemaSQL(ctx.Conf.SchemaName),
			schema.ID, schema.Version, schema.Filename, schema.Hash, schema.Status, schema.CreatedAt,
		)
		if err != nil {
			panic(fmt.Errorf("apply filename: %s, version: %s, insert schema error: %s",
				ms.migration.Filename, ms.migration.version, err,
			))
		}
	}

	migrator.ctx.Conf.Logger.Infof("%s", log.AnsiColorGreen("baseline successfully"))
	return nil
}

type upMigrator struct {
	ctx *Context
}

func NewUpMigrator(ctx *Context) (Migrator, error) {
	err := fillContext(ctx, true, false)
	if err != nil {
		return nil, err
	}
	ctx.Conf.printStatus = false
	statusMigrator := &statusMigrator{ctx}
	err = statusMigrator.Apply()
	return upMigrator{ctx}, err
}

func (migrator upMigrator) Apply() error {
	conf := migrator.ctx.Conf

	for _, ms := range migrator.ctx.ms {
		switch ms.status {
		case StatusHashMismatch:
			return fmt.Errorf("filename: %s, version %s, hash mismatch, expected: %s, actual: %s",
				ms.migration.Filename, ms.migration.version, ms.schema.Hash, ms.migration.fileHash,
			)
		case StatusFilenameMismatch:
			return fmt.Errorf("filename: %s, version %s, filename mismatch, expected: %s, actual: %s",
				ms.migration.Filename, ms.migration.version, ms.schema.Filename, ms.migration.Filename,
			)
		}
	}

	// apply migrations
	for _, ms := range migrator.ctx.ms {
		if ms.status == StatusBaseline || ms.status == StatusApplied {
			continue
		}
		if ms.status == StatusOutOfOrder && !migrator.ctx.Conf.MigrateOutOfOrder {
			conf.Logger.Warnf("filename: %s, version %s, out of order, skipped",
				ms.migration.Filename, ms.migration.version,
			)
		}

		var err error
		ms.schema, err = apply(migrator.ctx, ms.migration, true)
		if err != nil {
			return err
		}
	}

	migrator.ctx.Conf.Logger.Infof("%s", log.AnsiColorGreen("migration up successfully"))
	return nil
}

type downMigrator struct {
	ctx       *Context
	toVersion string
}

// NewDownMigrator creates a new down migrator
//
// toVersion specify the version that needs to be rolled back to. If it is empty,
// it will roll back to the first version (include this version). If it is not empty, it will roll back to this version
// (not including this version)
func NewDownMigrator(ctx *Context, toVersion string) (Migrator, error) {
	err := fillContext(ctx, false, true)
	if err != nil {
		return nil, err
	}
	ctx.Conf.printStatus = false
	statusMigrator := &statusMigrator{ctx}
	err = statusMigrator.Apply()
	return downMigrator{ctx, toVersion}, err
}

func (migrator downMigrator) Apply() error {
	targetVersionExists := migrator.toVersion == ""
	for _, ms := range migrator.ctx.ms {
		if !targetVersionExists {
			targetVersionExists = ms.migration.version == migrator.toVersion
		}
		switch ms.status {
		case StatusHashMismatch:
			return fmt.Errorf("filename: %s, version %s, hash mismatch, expected: %s, actual: %s",
				ms.migration.Filename, ms.migration.version, ms.schema.Hash, ms.migration.fileHash,
			)
		case StatusFilenameMismatch:
			return fmt.Errorf("filename: %s, version %s, filename mismatch, expected: %s, actual: %s",
				ms.migration.Filename, ms.migration.version, ms.schema.Filename, ms.migration.Filename,
			)
		}
	}
	if !targetVersionExists {
		return fmt.Errorf("no migration file with version %s found", migrator.toVersion)
	}

	// apply migrations
	for i := len(migrator.ctx.ms) - 1; i >= 0; i-- {
		ms := migrator.ctx.ms[i]
		if ms.status != StatusApplied {
			continue
		}
		if migrator.toVersion != "" && !CompareVersion(migrator.toVersion, ms.migration.version) {
			migrator.ctx.Conf.Logger.Debugf("filename: %s, version %s, skipped",
				ms.migration.Filename, ms.migration.version,
			)
			continue
		}

		_, err := apply(migrator.ctx, ms.migration, false)
		if err != nil {
			return err
		}
	}

	if !migrator.ctx.Conf.DryRun {
		migrator.ctx.Conf.Logger.Infof("%s", log.AnsiColorGreen("migration down successfully"))
	}
	return nil
}

func apply(ctx *Context, migration *Migration, up bool) (schema *Schema, err error) {
	statements := migration.upStatements
	if !up {
		statements = migration.downStatements
	}
	db := ctx.Conf.DB
	tx, err := db.BeginTx(ctx.Context, &sql.TxOptions{})

	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		} else {
			tx.Commit()
		}
	}()

	for _, statement := range statements {
		if ctx.Conf.DryRun {
			fmt.Println(string(statement))
			continue
		}
		_, err := tx.ExecContext(ctx.Context, string(statement))
		if err != nil {
			panic(fmt.Errorf("apply filename: %s, version: %s, statement: %s, error: %s",
				migration.Filename, migration.version, statement, err,
			))
		}
		ctx.Conf.Logger.Debugf("apply filename: %s, version: %s, statement: %s success",
			migration.Filename, migration.version, statement,
		)
	}
	ctx.Conf.Logger.Infof("apply filename: %s, version: %s success", migration.Filename, migration.version)

	if ctx.Conf.DryRun {
		return nil, nil
	}

	if up {
		ctx.Conf.schemaMaxID++
		schema = &Schema{
			ID:        ctx.Conf.schemaMaxID,
			Version:   migration.version,
			Filename:  migration.Filename,
			Hash:      migration.fileHash,
			Status:    StatusApplied,
			CreatedAt: time.Now(),
		}
		_, err = tx.ExecContext(
			ctx.Context,
			ctx.Conf.Dialect.InsertSchemaSQL(ctx.Conf.SchemaName),
			schema.ID, schema.Version, schema.Filename, schema.Hash, schema.Status, schema.CreatedAt,
		)
		if err != nil {
			panic(fmt.Errorf("apply filename: %s, version: %s, insert schema error: %s",
				migration.Filename, migration.version, err,
			))
		}
		return schema, nil
	}

	_, err = tx.ExecContext(
		ctx.Context,
		ctx.Conf.Dialect.DeleteSchemaSQL(ctx.Conf.SchemaName),
		migration.Filename,
	)
	if err != nil {
		panic(fmt.Errorf("apply filename: %s, version: %s, delete schema error: %s",
			migration.Filename, migration.version, err,
		))
	}
	return nil, nil
}

type statusMigrator struct {
	ctx *Context
}

func NewStatusMigrator(ctx *Context) (Migrator, error) {
	ctx.Conf.printStatus = true
	err := fillContext(ctx, false, false)
	return statusMigrator{ctx}, err
}

func (migrator statusMigrator) Apply() error {
	conf := migrator.ctx.Conf

	rows, err := conf.DB.Query(fmt.Sprintf(`
SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC
`, migrator.ctx.Conf.SchemaName,
	))
	if err != nil {
		return err
	}

	// read schemas from db
	schemas := make([]*Schema, 0)
	versionToSchema := make(map[string]*Schema, 0)
	for rows.Next() {
		schema := &Schema{}
		err := rows.Scan(&schema.ID, &schema.Version, &schema.Filename, &schema.Hash, &schema.Status, &schema.CreatedAt)
		if err != nil {
			return err
		}
		schemas = append(schemas, schema)
		versionToSchema[schema.Version] = schema

		migrator.ctx.Conf.schemaMaxID = schema.ID
	}
	rows.Close()
	sort.Slice(schemas, func(i, j int) bool {
		return CompareVersion(schemas[i].Version, schemas[j].Version)
	})

	// build status
	ms := make([]migrationStatus, 0)
	// | version | filename | hash | status |
	var (
		versionMaxLength  = len("Version")
		filenameMaxLength = len("Filename")
		hashMaxLength     = len("Hash")
		statusMaxLength   = len("Status")
	)
	for _, migration := range migrator.ctx.migrations {
		schema, ok := versionToSchema[migration.version]
		if !ok {
			status := StatusPending
			if len(schemas) > 0 && CompareVersion(migration.version, schemas[len(schemas)-1].Version) {
				status = StatusOutOfOrder
			}
			ms = append(ms, migrationStatus{
				migration: migration,
				schema:    nil,
				status:    status,
			})
		} else if schema.Filename != migration.Filename {
			ms = append(ms, migrationStatus{
				migration: migration,
				schema:    schema,
				status:    StatusFilenameMismatch,
			})
		} else if schema.Hash != migration.fileHash {
			ms = append(ms, migrationStatus{
				migration: migration,
				schema:    schema,
				status:    StatusHashMismatch,
			})
		} else {
			ms = append(ms, migrationStatus{
				migration: migration,
				schema:    schema,
				status:    schema.Status,
			})
		}

		{
			versionMaxLength = max(versionMaxLength, len(migration.version))
			filenameMaxLength = max(filenameMaxLength, len(migration.Filename))
			hashMaxLength = max(hashMaxLength, len(migration.fileHash))
			versionMaxLength = max(versionMaxLength, len(migration.version))
			statusMaxLength = max(statusMaxLength, len(ms[len(ms)-1].status))
		}
	}
	migrator.ctx.ms = ms

	// print status
	if !migrator.ctx.Conf.printStatus {
		return nil
	}

	header := fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s |",
		versionMaxLength, "Version",
		filenameMaxLength, "Filename",
		hashMaxLength, "Hash",
		statusMaxLength, "Status",
	)
	migrator.ctx.Conf.Logger.Infof(header)
	migrator.ctx.Conf.Logger.Infof("| %s + %s + %s + %s |",
		strings.Repeat("-", versionMaxLength),
		strings.Repeat("-", filenameMaxLength),
		strings.Repeat("-", hashMaxLength),
		strings.Repeat("-", statusMaxLength),
	)
	for _, m := range ms {
		migrator.ctx.Conf.Logger.Infof("| %-*s | %-*s | %-*s | %-*s |",
			versionMaxLength, m.migration.version,
			filenameMaxLength, m.migration.Filename,
			hashMaxLength, m.migration.fileHash,
			statusMaxLength, m.status,
		)
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
