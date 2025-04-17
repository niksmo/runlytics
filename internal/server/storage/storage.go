// Package storage provides memory and SQL database storage objects.
package storage

import (
	"database/sql"

	"github.com/niksmo/runlytics/pkg/di"
)

// New is a fabric method, returns Repository.
func New(db *sql.DB, fo di.FileOperator, config di.ServerConfig) di.Repository {
	if config.IsDatabase() {
		return NewPSQL(db)
	}
	return NewMemory(
		fo,
		config.SaveInterval(),
		config.Restore(),
	)
}
