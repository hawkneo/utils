package db

import (
	"database/sql"
	"fmt"

	"github.com/hawkneo/utils/web/health"
)

const (
	DEFAULT_QUERY = "SELECT 1"
)

var (
	_ health.Indicator = (*databaseIndicator)(nil)
)

type databaseIndicator struct {
	name  string
	query string
	db    *sql.DB
}

// NewDatabaseIndicator creates a new database indicator with the default query.
//
// name is the name of the indicator.
func NewDatabaseIndicator(name string, db *sql.DB) health.Indicator {
	return &databaseIndicator{
		name:  name,
		query: DEFAULT_QUERY,
		db:    db,
	}
}

// NewDatabaseIndicatorWithQuery creates a new database indicator with the given query.
//
// name is the name of the indicator.
func NewDatabaseIndicatorWithQuery(name string, db *sql.DB, query string) health.Indicator {
	return &databaseIndicator{
		name:  name,
		query: query,
		db:    db,
	}
}

func (d databaseIndicator) Name() string {
	return d.name
}

func (d databaseIndicator) Health() health.Health {
	rows, err := d.db.Query(d.query)
	if err != nil {
		return health.NewDownHealth(err)
	}
	defer rows.Close()

	for rows.Next() {
		return health.NewUpHealth()
	}
	return health.NewUnknownHealth(fmt.Errorf("no rows returned"))
}
