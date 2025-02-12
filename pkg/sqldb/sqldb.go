package sqldb

import (
	"database/sql"

	"github.com/niksmo/runlytics/pkg/di"
)

func New(driver, dsn string, logger di.Logger) *sql.DB {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		logger.Infow("SQL driver initialization", "error", err)
	}
	if err = db.Ping(); err != nil {
		logger.Debugw("Database ping", "error", err)
	}
	return db
}
