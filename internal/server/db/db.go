package db

import (
	"database/sql"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

func Init(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log.Error("Open DB", zap.Error(err))
	}
	if err = db.Ping(); err != nil {
		logger.Log.Info("DB not connected")
	} else {
		logger.Log.Info("DB connected")
	}
	return db
}
