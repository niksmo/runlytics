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

var tries = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

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
	go func() {
		defer wg.Done()
		<-stopCtx.Done()
		if err := ps.db.Close(); err != nil {
			logger.Log.Error("Database connection close error", zap.Error(err))
			return
		}
		logger.Log.Debug("Database disconnected")
	}()
}

func (ps *psqlStorage) Close() {
	ps.db.Close()
}

func (ps *psqlStorage) UpdateCounterByName(
	ctx context.Context, name string, value int64,
) (int64, error) {
	var (
		row *sql.Row
		err error
	)
	queryFn := func() error {
		row = ps.db.QueryRowContext(
			ctx,
			`INSERT INTO counter (name, value)

			VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET
			 value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value
			 RETURNING value;`,
			name,
			value,
		)
		return row.Err()
	}
	repeat.WithTries("Update counter by name query", tries, queryFn)
	if err != nil {
		return 0, err
	}

	var retValue int64
	queryFn = func() error {
		err = row.Scan(&retValue)
		return err
	}
	repeat.WithTries("Update counter by name scan", tries, queryFn)
	if err != nil {
		return 0, err
	}

	return retValue, err
}

func (ps *psqlStorage) UpdateGaugeByName(
	ctx context.Context, name string, value float64,
) (float64, error) {
	var (
		row *sql.Row
		err error
	)

	queryFn := func() error {
		row = ps.db.QueryRowContext(
			ctx,
			`INSERT INTO gauge (name, value)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET
			 value = EXCLUDED.value
			 RETURNING value;`,
			name,
			value,
		)
		return row.Err()
	}

	repeat.WithTries("Update gauge by name query", tries, queryFn)
	if err != nil {
		return 0, err
	}

	var retValue float64
	queryFn = func() error {
		err = row.Scan(&retValue)
		return err
	}
	repeat.WithTries("Update gauge by name scan", tries, queryFn)
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
	repeat.WithTries("Update counter list begin transaction", tries, queryFn)
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
				tries,
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
		tries,
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
		tries,
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
				tries,
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
		tries,
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
	var (
		row *sql.Row
		err error
	)

	queryFn := func() error {
		row = ps.db.QueryRowContext(
			ctx,
			`SELECT value
			 FROM counter
			 WHERE name = $1;`,
			name,
		)
		return row.Err()
	}
	repeat.WithTries("Read counter by name query", tries, queryFn)
	if err != nil {
		return 0, err
	}

	var value int64
	queryFn = func() error {
		err = row.Scan(&value)
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
			return nil
		}
		return err
	}
	repeat.WithTries("Read counter by name scan", tries, queryFn)
	if err != nil {
		return 0, err
	}

	return value, err
}

func (ps *psqlStorage) ReadGaugeByName(
	ctx context.Context, name string,
) (float64, error) {
	var (
		row *sql.Row
		err error
	)

	queryFn := func() error {
		row = ps.db.QueryRowContext(
			ctx,
			`SELECT value
			FROM gauge
			WHERE name = $1;`,
			name,
		)
		return row.Err()
	}
	repeat.WithTries("Read gauge by name query", tries, queryFn)
	if err != nil {
		return 0, err
	}

	var value float64
	queryFn = func() error {
		err = row.Scan(&value)
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
			return nil
		}
		return err
	}
	repeat.WithTries("Read gauge by name scan", tries, queryFn)
	if err != nil {
		return 0, err
	}

	return value, err
}

func (ps *psqlStorage) ReadGauge(
	ctx context.Context,
) (map[string]float64, error) {
	var (
		rows *sql.Rows
		err  error
	)

	queryFn := func() error {
		rows, err = ps.db.QueryContext(
			ctx, `SELECT name, value FROM gauge;`,
		)
		return err
	}

	repeat.WithTries("Read gauge", tries, queryFn)
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
	var (
		rows *sql.Rows
		err  error
	)

	queryFn := func() error {
		rows, err = ps.db.QueryContext(
			ctx, `SELECT name, value FROM counter;`,
		)
		return err
	}

	repeat.WithTries("Read counter", tries, queryFn)
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
	query := `
	CREATE TABLE IF NOT EXISTS gauge (
	    name TEXT PRIMARY KEY,
		value DOUBLE PRECISION NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS counter (
	    name TEXT PRIMARY KEY,
		value BIGINT NOT NULL
	);`

	queryFn := func() error {
		_, err := ps.db.ExecContext(ctx, query)
		return err
	}

	repeat.WithTries(
		"PSQL storage create tables",
		tries,
		queryFn,
	)
}
