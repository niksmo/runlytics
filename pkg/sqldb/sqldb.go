// Package sqldb provides [*sql.DB] constructor.
package sqldb

import (
	"database/sql"
)

// New returns sql.DB pointer.
func New(driver, dsn string, errFn func(error)) *sql.DB {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		errFn(err)
	}
	if err = db.Ping(); err != nil {
		errFn(err)
	}
	return db
}
