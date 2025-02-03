package storage

import (
	"context"
	"database/sql"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

type psqlStorage struct {
	db *sql.DB
}

func NewPSQL(db *sql.DB) *psqlStorage {
	return &psqlStorage{db: db}
}

func (ps *psqlStorage) Run() {
	ps.createTables()
}

func (ps *psqlStorage) UpdateCounterByName(name string, value int64) (int64, error) {
	row := ps.db.QueryRowContext(
		context.TODO(),
		`INSERT INTO counter (name, value)
		 VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET
		 value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value
		 RETURNING value;`,
		name,
		value,
	)

	var ret int64
	err := row.Scan(&ret)
	return ret, err
}

func (ps *psqlStorage) UpdateGaugeByName(name string, value float64) (float64, error) {
	row := ps.db.QueryRowContext(
		context.TODO(),
		`INSERT INTO gauge (name, value)
		 VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET
		 value = EXCLUDED.value
		 RETURNING value;`,
		name,
		value,
	)

	var ret float64
	err := row.Scan(&ret)
	return ret, err
}

func (ps *psqlStorage) ReadCounterByName(name string) (int64, error) {
	// ps.mu.RLock()
	// value, ok := ps.data.Counter[name]
	// ps.mu.RUnlock()

	// if !ok {
	// 	logger.Log.Debug("Not found counter metric", zap.String("name", name))
	// 	return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	// }
	// return value, nil
	return int64(1), nil
}

func (ps *psqlStorage) ReadGaugeByName(name string) (float64, error) {
	// ps.mu.RLock()
	// value, ok := ps.data.Gauge[name]
	// ps.mu.RUnlock()

	// if !ok {
	// 	logger.Log.Debug("Not found gauge metric", zap.String("name", name))
	// 	return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	// }
	// return value, nil
	return 1.0, nil
}

func (ps *psqlStorage) ReadGauge() map[string]float64 {
	// gauge := make(map[string]float64, len(ps.data.Gauge))

	// ps.mu.RLock()
	// for k, v := range ps.data.Gauge {
	// 	gauge[k] = v
	// }
	// ps.mu.RUnlock()

	// return gauge
	return make(map[string]float64)
}

func (ps *psqlStorage) ReadCounter() map[string]int64 {
	// counter := make(map[string]int64, len(ps.data.Counter))

	// ps.mu.RLock()
	// for k, v := range ps.data.Counter {
	// 	counter[k] = v
	// }
	// ps.mu.RUnlock()

	// return counter
	return make(map[string]int64)
}

func (ps *psqlStorage) createTables() {
	_, err := ps.db.ExecContext(context.TODO(), `
	CREATE TABLE IF NOT EXISTS gauge (
	    name TEXT PRIMARY KEY,
		value DOUBLE PRECISION NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS counter (
	    name TEXT PRIMARY KEY,
		value BIGINT NOT NULL
	);`)

	if err != nil {
		logger.Log.Error("Create tables", zap.Error(err))
	}
}
