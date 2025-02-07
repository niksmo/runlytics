package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

type psqlStorage struct {
	db *sql.DB
}

func newPSQL(db *sql.DB) *psqlStorage {
	return &psqlStorage{db}
}

// Creating database tables and waiting graceful shutdown
func (ps *psqlStorage) Run(stopCtx context.Context, wg *sync.WaitGroup) {
	ps.createTables()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-stopCtx.Done()
		if err := ps.db.Close(); err != nil {
			logger.Log.Error("Database connection close error", zap.Error(err))
			return
		}
		logger.Log.Debug("Database connection close properly")
	}()
}

func (ps *psqlStorage) Close() {
	ps.db.Close()
}

func (ps *psqlStorage) UpdateCounterByName(
	ctx context.Context, name string, value int64,
) (int64, error) {
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

func (ps *psqlStorage) UpdateCounterList(ctx context.Context, m map[string]int64) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for name, value := range m {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO counter (name, value)
		     VALUES ($1, $2)
		     ON CONFLICT (name) DO UPDATE SET
		     value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value;`,
			name,
			value,
		)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}
	return tx.Commit()
}

func (ps *psqlStorage) UpdateGaugeList(ctx context.Context, m map[string]float64) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for name, value := range m {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO gauge (name, value)
		    VALUES ($1, $2)
		    ON CONFLICT (name) DO UPDATE SET
		    value = EXCLUDED.value;`,
			name,
			value,
		)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}
	return tx.Commit()
}

func (ps *psqlStorage) ReadCounterByName(
	ctx context.Context, name string,
) (int64, error) {
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
