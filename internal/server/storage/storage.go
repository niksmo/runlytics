package storage

import (
	"database/sql"

	"github.com/niksmo/runlytics/pkg/di"
)

func New(db *sql.DB, config di.Config) di.Repository {
	if config.IsDatabase() {
		return newPSQL(db)
	}
	return newMemory(
		config.File(),
		config.SaveInterval(),
		config.Restore(),
	)
}
