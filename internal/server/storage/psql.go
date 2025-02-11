package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/repeat"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

var waitIntervals = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second, 5 * time.Second, 5 * time.Second}

type psqlStorage struct {
	db *sql.DB
}

func newPSQL(db *sql.DB) *psqlStorage {
	return &psqlStorage{db}
}

// Creating database tables and waiting graceful shutdown
func (ps *psqlStorage) Run(stopCtx context.Context, wg *sync.WaitGroup) {
	ps.createTables(stopCtx)

	wg.Add(1)
	go ps.waitStop(stopCtx, wg)
}

func (ps *psqlStorage) Close() error {
	return ps.db.Close()
}

func (ps *psqlStorage) UpdateCounterByName(
	ctx context.Context, name string, value int64,
) (int64, error) {
	stmt := `INSERT INTO counter (name, value)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET
			 value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value
			 RETURNING value;`
	row := ps.db.QueryRowContext(ctx, stmt, name, value)

	var retValue int64
	err := scanRowWithRetries(ctx, row, "Update counter by", &retValue)
	if err != nil {
		return 0, err
	}

	return retValue, err
}

func (ps *psqlStorage) UpdateGaugeByName(
	ctx context.Context, name string, value float64,
) (float64, error) {
	stmt := `INSERT INTO gauge (name, value)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET
			 value = EXCLUDED.value
			 RETURNING value;`
	row := ps.db.QueryRowContext(ctx, stmt, name, value)

	var retValue float64
	err := scanRowWithRetries(ctx, row, "Update gauge by name", &retValue)
	if err != nil {
		return 0, err
	}

	return retValue, err
}

func (ps *psqlStorage) UpdateCounterList(
	ctx context.Context, mSlice []metrics.MetricsCounter,
) error {

	var (
		tx  *sql.Tx
		err error
	)

	queryFn := func() error {
		tx, err = ps.db.BeginTx(ctx, nil)
		return err
	}
	repeat.WithTries("Update counter list begin transaction", waitIntervals, queryFn)
	if err != nil {
		return err
	}

	for _, item := range mSlice {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO counter (name, value)
		     VALUES ($1, $2)
		     ON CONFLICT (name) DO UPDATE SET
		     value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value;`,
			item.ID,
			item.Delta,
		)
		if err != nil {
			repeat.WithTries(
				"Update counter list rollback",
				waitIntervals,
				func() error {
					err = tx.Rollback()
					return err
				},
			)
			return err
		}
	}

	repeat.WithTries(
		"Update counter list commit",
		waitIntervals,
		func() error {
			err = tx.Commit()
			return err
		},
	)
	return err
}

func (ps *psqlStorage) UpdateGaugeList(
	ctx context.Context, mSlice []metrics.MetricsGauge,
) error {
	var (
		tx  *sql.Tx
		err error
	)

	repeat.WithTries(
		"Update gauge list begin transaction",
		waitIntervals,
		func() error {
			tx, err = ps.db.BeginTx(ctx, nil)
			return err
		},
	)
	if err != nil {
		return err
	}

	for _, item := range mSlice {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO gauge (name, value)
		    VALUES ($1, $2)
		    ON CONFLICT (name) DO UPDATE SET
		    value = EXCLUDED.value;`,
			item.ID,
			item.Value,
		)
		if err != nil {
			repeat.WithTries(
				"Update gauge list rollback",
				waitIntervals,
				func() error {
					err = tx.Rollback()
					return err
				},
			)
			return err
		}
	}

	repeat.WithTries(
		"Update gauge list commit",
		waitIntervals,
		func() error {
			err = tx.Commit()
			return err
		},
	)
	return err
}

func (ps *psqlStorage) ReadCounterByName(
	ctx context.Context, name string,
) (int64, error) {
	stmt := `SELECT value FROM counter WHERE name = $1;`
	row := ps.db.QueryRow(stmt, name)

	var value int64
	err := scanRowWithRetries(ctx, row, "Read counter by name", &value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
		}
		return 0, err
	}

	return value, err
}

func (ps *psqlStorage) ReadGaugeByName(
	ctx context.Context, name string,
) (float64, error) {
	stmt := `SELECT value FROM gauge WHERE name = $1;`
	row := ps.db.QueryRow(stmt, name)

	var value float64
	err := scanRowWithRetries(ctx, row, "Read gauge by name", &value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
		}
		return 0, err
	}

	return value, err
}

func (ps *psqlStorage) ReadGauge(
	ctx context.Context,
) (map[string]float64, error) {
	stmt := `SELECT name, value FROM gauge;`
	rows, err := queryWithRetries(ctx, ps.db, stmt, "Read gauge")
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
	stmt := `SELECT name, value FROM counter;`
	rows, err := queryWithRetries(ctx, ps.db, stmt, "Read counter")
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

func (ps *psqlStorage) createTables(ctx context.Context) {
	stmt := `
	CREATE TABLE IF NOT EXISTS gauge (
	    name TEXT PRIMARY KEY,
		value DOUBLE PRECISION NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS counter (
	    name TEXT PRIMARY KEY,
		value BIGINT NOT NULL
	);`

	_, err := execWithRetries(ctx, ps.db, stmt, "Create tables")
	if err != nil {
		logger.Log.Error("Crate tables", zap.Error(err))
	}
}

func (ps *psqlStorage) waitStop(stopCtx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	<-stopCtx.Done()
	if err := ps.Close(); err != nil {
		logger.Log.Error("Database connection close error", zap.Error(err))
		return
	}
	logger.Log.Debug("Database disconnected")
}

const tryAfter = "tryAfter"

func execWithRetries(
	ctx context.Context, db *sql.DB, stmt, logPrefix string, args ...any,
) (sql.Result, error) {
	result, err := db.ExecContext(ctx, stmt, args...)
	if err != nil {
	retries:
		for _, interval := range waitIntervals {
			logger.Log.Debug(logPrefix, zap.Duration(tryAfter, interval))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(interval):
				result, err = db.ExecContext(ctx, stmt, args...)
				if err == nil {
					break retries
				}
			}
		}
	}
	return result, err
}

func queryWithRetries(
	ctx context.Context, db *sql.DB, stmt, logPrefix string, args ...any,
) (*sql.Rows, error) {
	rows, err := db.QueryContext(ctx, stmt, args...)

	if err != nil {
	retries:
		for _, interval := range waitIntervals {
			logger.Log.Debug(
				logPrefix, zap.Duration(tryAfter, interval), zap.Error(err),
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(interval):
				rows, err = db.QueryContext(ctx, stmt, args...)
				if err == nil {
					break retries
				}
			}
		}
	}

	return rows, err
}

func scanRowWithRetries(
	ctx context.Context, row *sql.Row, logPrefix string, args ...any,
) error {
	err := row.Scan(args...)
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if err != nil {
	retries:
		for _, interval := range waitIntervals {
			logger.Log.Info(
				logPrefix,
				zap.Duration(tryAfter, interval),
				zap.Error(err),
			)
			select {
			case <-ctx.Done():
				return err
			case <-time.After(interval):
				err = row.Scan(args...)
				if err == nil || errors.Is(err, sql.ErrNoRows) {
					break retries
				}
			}
		}
	}

	return err
}
