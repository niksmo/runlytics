package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPSQL(t *testing.T) {
	db, err := sql.Open(
		"pgx",
		os.Getenv("RUNLYTICS_TEST_DSN"),
	)
	require.NoError(t, err)
	defer db.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Second)
	defer pingCancel()

	err = db.PingContext(pingCtx)
	if err != nil {
		t.SkipNow()
		return
	}
	require.NoError(t, err)

	clearTables := func(t *testing.T) {
		_, err := db.ExecContext(
			context.TODO(),
			`TRUNCATE TABLE gauge, counter;`,
		)
		require.NoError(t, err)
	}

	t.Run("Create tables on running", func(t *testing.T) {
		ctxBase := context.Background()

		arrangeCtx0, arrangeCancel0 := context.WithTimeout(ctxBase, time.Second)
		defer arrangeCancel0()
		_, err := db.ExecContext(
			arrangeCtx0,
			`DROP TABLE IF EXISTS gauge;
		     DROP TABLE IF EXISTS counter;`,
		)
		require.NoError(t, err)

		qCountTables := `
		SELECT COUNT(tablename)
			FROM pg_catalog.pg_tables
			WHERE tablename IN ('gauge', 'counter');
		`

		arrangeCtx1, arrangeCancel1 := context.WithTimeout(ctxBase, time.Second)
		defer arrangeCancel1()
		row := db.QueryRowContext(
			arrangeCtx1,
			qCountTables,
		)
		var count int
		require.NoError(t, row.Scan(&count))
		assert.Zero(t, count)

		storage := NewPSQL(db)
		storage.Run()

		actCtx, actCancel := context.WithTimeout(ctxBase, time.Second)
		defer actCancel()

		row = db.QueryRowContext(actCtx, qCountTables)
		require.NoError(t, row.Scan(&count))
		assert.Equal(t, 2, count)
	})

	t.Run("Sequence update gauge by name", func(t *testing.T) {
		clearTables(t)
		storage := NewPSQL(db)
		storage.Run()
		ctxBase := context.Background()
		metricName := "Alloc"
		seq := []float64{77.55, 33.22, 0}
		for _, v := range seq {
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			actualValue, err := storage.UpdateGaugeByName(
				ctx, metricName, v,
			)
			require.NoError(t, err)
			assert.Equal(t, v, actualValue)
		}
	})

	t.Run("Sequence update counter by name", func(t *testing.T) {
		clearTables(t)
		storage := NewPSQL(db)
		storage.Run()
		ctxBase := context.Background()
		metricName := "Counter"
		seq := []int64{5, 5, 5}
		var sum int64
		for _, v := range seq {
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			actualValue, err := storage.UpdateCounterByName(
				ctx, metricName, v,
			)
			require.NoError(t, err)
			assert.Equal(t, sum+v, actualValue)
			sum += v
		}
	})

	t.Run("Read counter by name", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run()
		ctxBase := context.Background()
		metricName := "Counter"

		t.Run("Should return NoExists error", func(t *testing.T) {
			clearTables(t)
			expected := int64(0)
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			actualValue, err := storage.ReadCounterByName(
				ctx, metricName,
			)
			require.ErrorIs(t, err, server.ErrNotExists)
			assert.Equal(t, expected, actualValue)
		})

		t.Run("Should return entry value", func(t *testing.T) {
			clearTables(t)
			expected := int64(10)
			arrangeCtx, arrangeCancel := context.WithTimeout(ctxBase, time.Second)
			defer arrangeCancel()
			_, err := db.ExecContext(
				arrangeCtx,
				`INSERT INTO counter (name, value)
				 VALUES ($1, $2);`,
				metricName,
				expected,
			)
			require.NoError(t, err)
			actCtx, actCancel := context.WithTimeout(ctxBase, time.Second)
			defer actCancel()
			actualValue, err := storage.ReadCounterByName(
				actCtx, metricName,
			)
			require.NoError(t, err)
			assert.Equal(t, expected, actualValue)
		})
	})

	t.Run("Read gauge by name", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run()
		ctxBase := context.Background()
		metricName := "Alloc"

		t.Run("Should return NotExists error", func(t *testing.T) {
			clearTables(t)
			expected := float64(0)
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			actualValue, err := storage.ReadGaugeByName(
				ctx, metricName,
			)
			require.ErrorIs(t, err, server.ErrNotExists)
			assert.Equal(t, expected, actualValue)
		})

		t.Run("Should return entry value", func(t *testing.T) {
			clearTables(t)
			expected := float64(10)
			arrangeCtx, arrangeCancel := context.WithTimeout(ctxBase, time.Second)
			defer arrangeCancel()
			_, err := db.ExecContext(
				arrangeCtx,
				`INSERT INTO gauge (name, value)
				 VALUES ($1, $2);`,
				metricName,
				expected,
			)
			require.NoError(t, err)
			actCtx, actCancel := context.WithTimeout(ctxBase, time.Second)
			defer actCancel()
			actualValue, err := storage.ReadGaugeByName(
				actCtx, metricName,
			)
			require.NoError(t, err)
			assert.Equal(t, expected, actualValue)
		})
	})

	t.Run("Read counter", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run()
		ctxBase := context.Background()

		t.Run("Should return empty data", func(t *testing.T) {
			clearTables(t)
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			expectedDataLen := 0
			counterData, err := storage.ReadCounter(ctx)
			require.NoError(t, err)
			assert.Len(t, counterData, expectedDataLen)
		})

		t.Run("Should return data", func(t *testing.T) {
			clearTables(t)
			testList := []struct {
				name  string
				value int64
			}{
				{"Counter0", int64(10)}, {"Counter1", int64(15)},
			}
			expectedDataLen := len(testList)

			insertValues := make([]string, 0, expectedDataLen)
			for _, counter := range testList {
				insertValues = append(
					insertValues,
					fmt.Sprintf("('%s', %v)", counter.name, counter.value),
				)
			}

			arrangeCtx, arrangeCancel := context.WithTimeout(ctxBase, time.Second)
			defer arrangeCancel()
			_, err := db.ExecContext(
				arrangeCtx,
				`INSERT INTO counter (name, value) VALUES `+
					strings.Join(insertValues, ", ")+";",
			)
			require.NoError(t, err)

			actCtx, actCancel := context.WithTimeout(ctxBase, time.Second)
			defer actCancel()
			counterData, err := storage.ReadCounter(actCtx)
			require.NoError(t, err)
			assert.Len(t, counterData, expectedDataLen)
			for _, counter := range testList {
				assert.Equal(t, counter.value, counterData[counter.name])
			}
		})
	})

	t.Run("Read gauge", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run()
		ctxBase := context.Background()

		t.Run("Should return empty data", func(t *testing.T) {
			clearTables(t)
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			expectedDataLen := 0
			counterData, err := storage.ReadGauge(ctx)
			require.NoError(t, err)
			assert.Len(t, counterData, expectedDataLen)
		})

		t.Run("Should return data", func(t *testing.T) {
			clearTables(t)
			testList := []struct {
				name  string
				value float64
			}{
				{"Gauge0", float64(10.55)}, {"Gauge1", float64(15.22)},
			}
			expectedDataLen := len(testList)

			insertValues := make([]string, 0, expectedDataLen)
			for _, gauge := range testList {
				insertValues = append(
					insertValues,
					fmt.Sprintf("('%s', %v)", gauge.name, gauge.value),
				)
			}

			arrangeCtx, arrangeCancel := context.WithTimeout(ctxBase, time.Second)
			defer arrangeCancel()
			_, err := db.ExecContext(
				arrangeCtx,
				`INSERT INTO gauge (name, value) VALUES `+
					strings.Join(insertValues, ", ")+";",
			)
			require.NoError(t, err)

			actCtx, actCancel := context.WithTimeout(ctxBase, time.Second)
			defer actCancel()
			gaugeData, err := storage.ReadGauge(actCtx)
			require.NoError(t, err)
			assert.Len(t, gaugeData, expectedDataLen)
			for _, gauge := range testList {
				assert.Equal(t, gauge.value, gaugeData[gauge.name])
			}
		})
	})
}
