package storage

import (
	"database/sql"

	"github.com/niksmo/runlytics/pkg/di"
)

func New(db *sql.DB, fo di.FileOperator, config di.ServerConfig) di.Repository {
	if config.IsDatabase() {
		return newPSQL(db)
	}
	return newMemory(
		fo,
		config.SaveInterval(),
		config.Restore(),
	)
}
