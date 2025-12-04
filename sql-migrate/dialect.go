package migrate

import (
	"fmt"
	"strings"
)

var (
	_ Dialect = (*PostgresDialect)(nil)
	_ Dialect = (*MySQLDialect)(nil)
)

type Dialect interface {
	CreateSchemaSQL(schemaName string) string
	InsertSchemaSQL(schemaName string) string
	DeleteSchemaSQL(schemaName string) string
}

type PostgresDialect struct {
}

func (PostgresDialect) CreateSchemaSQL(schemaName string) string {
	return strings.TrimSpace(fmt.Sprintf(`
CREATE TABLE %s
(
    id BIGSERIAL PRIMARY KEY NOT NULL ,
    version TEXT NOT NULL ,
    filename TEXT UNIQUE NOT NULL ,
    hash TEXT NOT NULL ,
    status TEXT NOT NULL ,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`, schemaName,
	))
}

func (PostgresDialect) InsertSchemaSQL(schemaName string) string {
	return fmt.Sprintf(`
INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)
`, schemaName)
}

func (PostgresDialect) DeleteSchemaSQL(schemaName string) string {
	return fmt.Sprintf(`
DELETE FROM %s WHERE filename = $1
`, schemaName)
}

type MySQLDialect struct{}

func (MySQLDialect) CreateSchemaSQL(schemaName string) string {
	return strings.TrimSpace(fmt.Sprintf(`
CREATE TABLE %s (
    id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    version BIGINT NOT NULL DEFAULT 0,
    filename VARCHAR(200) NOT NULL UNIQUE ,
    hash VARCHAR(100) NOT NULL ,
    status VARCHAR(20) NOT NULL ,
    created_at DATETIME NOT NULL DEFAULT NOW()
) CHARACTER SET utf8mb4
`, schemaName,
	))
}

func (MySQLDialect) InsertSchemaSQL(schemaName string) string {
	return fmt.Sprintf(`
INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)
`, schemaName)
}

func (MySQLDialect) DeleteSchemaSQL(schemaName string) string {
	return fmt.Sprintf(`
DELETE FROM %s WHERE filename = ?
`, schemaName)
}
