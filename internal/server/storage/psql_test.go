package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPSQL(t *testing.T) {
	db, err := sql.Open(
		"pgx",
		"postgres://runlytics:runlytics@127.0.0.1:5432/runlytics_tests"+
			"?sslmode=disable",
	)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(t, err)

	t.Run("Create tables on running", func(t *testing.T) {
		db.ExecContext(context.TODO(), `
		DROP TABLE IF EXISTS gauge;
		DROP TABLE IF EXISTS counter;
		`)

		qCountTables := `
		SELECT COUNT(tablename)
			FROM pg_catalog.pg_tables
			WHERE tablename IN ('gauge', 'counter');
		`

		row := db.QueryRowContext(context.TODO(), qCountTables)
		var count int
		require.NoError(t, row.Scan(&count))
		assert.Zero(t, count)

		storage := NewPSQL(db)
		storage.Run()

		row = db.QueryRowContext(context.TODO(), qCountTables)
		require.NoError(t, row.Scan(&count))
		assert.Equal(t, 2, count)
	})
}
