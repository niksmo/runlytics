// Package storage provides memory and SQL database storage objects.
package storage

import (
	"time"

	"github.com/niksmo/runlytics/internal/server/storage/filestorage"
	psqlstorage "github.com/niksmo/runlytics/internal/server/storage/postgresql"
	"github.com/niksmo/runlytics/pkg/di"
)

// New is a fabric method that returns storage with [di.Storage] interface.
func New(
	fo di.FileOperator, dsn string, saveInterval time.Duration, restore bool,
) di.IStorage {
	if dsn != "" {
		return psqlstorage.New(dsn)
	}

	return filestorage.New(fo, saveInterval, restore)
}
