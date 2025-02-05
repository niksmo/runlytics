package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

const (
	queryTimeout = time.Second
)

type psqlStorage struct {
	db           *sql.DB
	queryTimeout time.Duration
}

func NewPSQL(db *sql.DB) *psqlStorage {
	return &psqlStorage{db: db, queryTimeout: queryTimeout}
}

func (ps *psqlStorage) Run() {
	ps.createTables()
}

func (ps *psqlStorage) UpdateCounterByName(
	ctx context.Context, name string, value int64,
) (int64, error) {
	ctx, cancel := ps.newContext(ctx)
	defer cancel()

	row := ps.db.QueryRowContext(
		ctx,
		`INSERT INTO counter (name, value)
		 VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET
		 value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value
		 RETURNING value;`,
		name,
		value,
	)

	var actualValue int64
	err := row.Scan(&actualValue)
	return actualValue, err
}

func (ps *psqlStorage) UpdateGaugeByName(
	ctx context.Context, name string, value float64,
) (float64, error) {
	ctx, cancel := ps.newContext(ctx)
	defer cancel()

	row := ps.db.QueryRowContext(
		ctx,
		`INSERT INTO gauge (name, value)
		 VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET
		 value = EXCLUDED.value
		 RETURNING value;`,
		name,
		value,
	)

	var actualValue float64
	err := row.Scan(&actualValue)
	return actualValue, err
}

func (ps *psqlStorage) ReadCounterByName(
	ctx context.Context, name string,
) (int64, error) {
	ctx, cancel := ps.newContext(ctx)
	defer cancel()

	row := ps.db.QueryRowContext(
		ctx,
		`SELECT value
		 FROM counter
		 WHERE name = $1;`,
		name,
	)
	var value int64
	err := row.Scan(&value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
		}

		return 0, err
	}

	return value, err
}

func (ps *psqlStorage) ReadGaugeByName(
	ctx context.Context, name string,
) (float64, error) {
	ctx, cancel := ps.newContext(ctx)
	defer cancel()

	row := ps.db.QueryRowContext(
		ctx,
		`SELECT value
		 FROM gauge
		 WHERE name = $1;`,
		name,
	)
	var value float64
	err := row.Scan(&value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
		}

		return 0, err
	}

	return value, err
}

func (ps *psqlStorage) ReadGauge(
	ctx context.Context,
) (map[string]float64, error) {
	ctx, cancel := ps.newContext(ctx)
	defer cancel()

	rows, err := ps.db.QueryContext(
		ctx, `SELECT name, value FROM gauge;`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gaugeMap := make(map[string]float64)
	var (
		name  string
		value float64
	)
	for rows.Next() {
		if err = rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		gaugeMap[name] = value
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return gaugeMap, nil
}

func (ps *psqlStorage) ReadCounter(
	ctx context.Context,
) (map[string]int64, error) {
	ctx, cancel := ps.newContext(ctx)
	defer cancel()

	rows, err := ps.db.QueryContext(
		ctx, `SELECT name, value FROM counter;`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counterMap := make(map[string]int64)
	var (
		name  string
		value int64
	)
	for rows.Next() {
		if err = rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		counterMap[name] = value
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return counterMap, nil
}

func (ps *psqlStorage) newContext(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, ps.queryTimeout)
}

func (ps *psqlStorage) createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS gauge (
	    name TEXT PRIMARY KEY,
		value DOUBLE PRECISION NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS counter (
	    name TEXT PRIMARY KEY,
		value BIGINT NOT NULL
	);`

	if _, err := ps.db.ExecContext(context.TODO(), query); err != nil {
		logger.Log.Error("Create tables", zap.Error(err))
	}
}
