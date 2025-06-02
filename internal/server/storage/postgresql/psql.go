package psqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

var waitIntervals = []time.Duration{time.Second, 3 * time.Second}

// PSQLStorage wrap [*sql.DB] and implements [di.Storage] interface.
type PSQLStorage struct {
	db *sql.DB
}

// New returns PSQLStorage pointer.
func New(dsn string) *PSQLStorage {
	db, _ := sql.Open("pgx", dsn)
	return &PSQLStorage{db}
}

func (ps *PSQLStorage) MustRun() {
	if err := ps.Run(); err != nil {
		panic(err)
	}
}

// Creating database tables and waiting graceful shutdown
func (ps *PSQLStorage) Run() error {
	const op = "psqlstorage.Run"

	err := ps.checkDB(context.Background())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = ps.createTables()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (ps *PSQLStorage) Stop() {
	const op = "psqlstorage.Stop"
	if err := ps.db.Close(); err != nil {
		logger.Log.Error(
			"failed to close storage",
			zap.String("op", op), zap.Error(err),
		)
	}
}

func (ps *PSQLStorage) Ping(ctx context.Context) error {
	return ps.checkDB(ctx)
}

// UpdateCounterByName returns updated counter value and sql driver error, if occur.
func (ps *PSQLStorage) UpdateCounterByName(
	ctx context.Context, name string, value int64,
) (int64, error) {
	logPrefix := "Update counter by name"
	stmt := `INSERT INTO counter (name, value)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET
			 value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value
			 RETURNING value;`
	row := ps.db.QueryRowContext(ctx, stmt, name, value)

	var retValue int64
	err := scanRowWithRetries(ctx, row, logPrefix, &retValue)
	if err != nil {
		logger.Log.Error(logPrefix+": scan row", zap.Error(err))
		return 0, err
	}

	return retValue, nil
}

// UpdateGaugeByName returns updated gauge value and sql driver error, if occur.
func (ps *PSQLStorage) UpdateGaugeByName(
	ctx context.Context, name string, value float64,
) (float64, error) {
	logPrefix := "Update gauge by name"
	stmt := `INSERT INTO gauge (name, value)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET
			 value = EXCLUDED.value
			 RETURNING value;`
	row := ps.db.QueryRowContext(ctx, stmt, name, value)

	var retValue float64
	err := scanRowWithRetries(ctx, row, "Update gauge by name", &retValue)
	if err != nil {
		logger.Log.Error(logPrefix+": scan row", zap.Error(err))
		return 0, err
	}

	return retValue, nil
}

// UpdateCounterList returns sql driver error, if occur.
func (ps *PSQLStorage) UpdateCounterList(
	ctx context.Context, mSlice metrics.MetricsList,
) error {
	logPrefix := "Update counter list"
	tx, err := beginTxWithRetries(
		ctx, ps.db, logPrefix+": begin transaction", nil,
	)
	if err != nil {
		logger.Log.Error(logPrefix+": begin transaction", zap.Error(err))
		return err
	}
	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO counter (name, value)
		 VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET
		 value = (SELECT value FROM counter WHERE name=$1) + EXCLUDED.value;`,
	)
	if err != nil {
		logger.Log.Error(logPrefix+": prepare", zap.Error(err))
		return err
	}
	defer stmt.Close()

	for _, item := range mSlice {
		_, err = stmt.ExecContext(ctx, item.ID, item.Delta)
		if err != nil {
			err = rollbackWithRetries(ctx, tx, logPrefix+": rollback")
			if err != nil {
				logger.Log.Error(logPrefix+": rollback", zap.Error(err))
				return err
			}
		}
	}

	if err = commitWithRetries(ctx, tx, logPrefix+": commit"); err != nil {
		logger.Log.Error(logPrefix+": commit", zap.Error(err))
		return err
	}
	return nil
}

// UpdateGaugeList returns sql driver error, if occur.
func (ps *PSQLStorage) UpdateGaugeList(
	ctx context.Context, mSlice metrics.MetricsList,
) error {
	logPrefix := "Update gauge list"
	tx, err := beginTxWithRetries(
		ctx, ps.db, logPrefix+": begin transaction", nil,
	)
	if err != nil {
		logger.Log.Error(logPrefix+": begin transaction", zap.Error(err))
		return err
	}
	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO gauge (name, value)
	     VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET
		 value = EXCLUDED.value;`,
	)
	if err != nil {
		logger.Log.Error(logPrefix+": prepare", zap.Error(err))
		return err
	}
	defer stmt.Close()

	for _, item := range mSlice {
		_, err = stmt.ExecContext(ctx, item.ID, item.Value)
		if err != nil {
			err = rollbackWithRetries(ctx, tx, logPrefix+": rollback")
			if err != nil {
				logger.Log.Error(logPrefix+": rollback", zap.Error(err))
				return err
			}
		}
	}

	if err = commitWithRetries(ctx, tx, logPrefix+": commit"); err != nil {
		logger.Log.Error(logPrefix+": commit", zap.Error(err))
		return err
	}
	return nil
}

// ReadCounterByName returns counter value and sql driver error, if occur.
func (ps *PSQLStorage) ReadCounterByName(
	ctx context.Context, name string,
) (int64, error) {
	logPrefix := "Read counter by name"
	stmt := `SELECT value FROM counter WHERE name = $1;`
	row := ps.db.QueryRow(stmt, name)

	var value int64
	err := scanRowWithRetries(ctx, row, logPrefix, &value)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
	case err != nil:
		logger.Log.Error(logPrefix+": scan row", zap.Error(err))
		return 0, err
	}

	return value, nil
}

// ReadGaugeByName returns gauge value and sql driver error, if occur.
func (ps *PSQLStorage) ReadGaugeByName(
	ctx context.Context, name string,
) (float64, error) {
	logPrefix := "Read gauge by name"
	stmt := `SELECT value FROM gauge WHERE name = $1;`
	row := ps.db.QueryRow(stmt, name)

	var value float64
	err := scanRowWithRetries(ctx, row, logPrefix, &value)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
	case err != nil:
		logger.Log.Error(logPrefix+": scan row", zap.Error(err))
		return 0, err
	}

	return value, nil
}

// ReadGauge returns gauge metrics and sql driver error, if occur.
func (ps *PSQLStorage) ReadGauge(
	ctx context.Context,
) (map[string]float64, error) {
	logPrefix := "Read gauge"
	stmt := `SELECT name, value FROM gauge;`
	rows, err := queryWithRetries(ctx, ps.db, stmt, logPrefix)
	if err != nil {
		logger.Log.Error(logPrefix+": query", zap.Error(err))
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
			logger.Log.Error(logPrefix+": scan rows", zap.Error(err))
			return nil, err
		}
		gaugeMap[name] = value
	}
	if err = rows.Err(); err != nil {
		logger.Log.Error(logPrefix+": after rows scan iteration", zap.Error(err))
		return nil, err
	}

	return gaugeMap, nil
}

// ReadCounter returns counter metrics and sql driver error if occurs.
func (ps *PSQLStorage) ReadCounter(
	ctx context.Context,
) (map[string]int64, error) {
	logPrefix := "Read counter"
	stmt := `SELECT name, value FROM counter;`
	rows, err := queryWithRetries(ctx, ps.db, stmt, logPrefix)
	if err != nil {
		logger.Log.Error(logPrefix+": query", zap.Error(err))
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
			logger.Log.Error(logPrefix+": scan rows", zap.Error(err))
			return nil, err
		}
		counterMap[name] = value
	}
	if err = rows.Err(); err != nil {
		logger.Log.Error(logPrefix+": after rows scan iteration", zap.Error(err))
		return nil, err
	}

	return counterMap, nil
}

func (ps *PSQLStorage) checkDB(ctx context.Context) error {
	return ps.db.PingContext(ctx)
}

func (ps *PSQLStorage) createTables() error {
	const op = "psqlstorage.createTables"

	log := logger.Log.With(zap.String("op", op))

	stmt := `
	CREATE TABLE IF NOT EXISTS gauge (
	    name TEXT PRIMARY KEY,
		value DOUBLE PRECISION NOT NULL
	);
	CREATE TABLE IF NOT EXISTS counter (
	    name TEXT PRIMARY KEY,
		value BIGINT NOT NULL
	);`

	_, err := execWithRetries(context.Background(), ps.db, stmt, log)
	if err != nil {
		return err
	}
	return nil
}

const tryAfter = "tryAfter"

func execWithRetries(
	ctx context.Context, db *sql.DB, stmt string, log *zap.Logger, args ...any,
) (sql.Result, error) {
	result, err := db.ExecContext(ctx, stmt, args...)
	if err != nil {
	retries:
		for _, interval := range waitIntervals {
			log.Debug("", zap.Duration(tryAfter, interval))
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

func beginTxWithRetries(
	ctx context.Context, db *sql.DB, logPrefix string, opts *sql.TxOptions,
) (*sql.Tx, error) {
	tx, err := db.BeginTx(ctx, opts)
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
				return nil, ctx.Err()
			case <-time.After(interval):
				tx, err = db.BeginTx(ctx, opts)
				if err == nil {
					break retries
				}
			}
		}
	}

	return tx, err
}

func rollbackWithRetries(
	ctx context.Context, tx *sql.Tx, logPrefix string,
) error {
	err := tx.Rollback()
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
				return ctx.Err()
			case <-time.After(interval):
				err = tx.Rollback()
				if err == nil {
					break retries
				}
			}
		}
	}
	return err
}

func commitWithRetries(
	ctx context.Context, tx *sql.Tx, logPrefix string,
) error {
	err := tx.Commit()
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
				return ctx.Err()
			case <-time.After(interval):
				err = tx.Commit()
				if err == nil {
					break retries
				}
			}
		}
	}
	return err
}
