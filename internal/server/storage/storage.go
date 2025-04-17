// Package storage provides memory and SQL database storage objects.
package storage

import (
	"database/sql"

	"github.com/niksmo/runlytics/pkg/di"
)

// New is a fabric method that returns storage with [di.Repository] interface.
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
