package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPSQL(t *testing.T) {
	DSN := os.Getenv("RUNLYTICS_TEST_DSN")
	db, err := sql.Open(
		"pgx",
		DSN,
	)
	require.NoError(t, err)
	defer db.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Second)
	defer pingCancel()

	err = db.PingContext(pingCtx)
	if err != nil {
		t.Skip("Test database not connected")
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

	stopCtx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM,
	)
	defer stop()
	var wg sync.WaitGroup

	t.Run("Create tables on running", func(t *testing.T) {
		ctxBase := context.Background()

		ctx, cancel := context.WithTimeout(ctxBase, time.Second)
		defer cancel()
		_, err := db.ExecContext(
			ctx,
			`DROP TABLE IF EXISTS gauge;
		     DROP TABLE IF EXISTS counter;`,
		)
		require.NoError(t, err)

		qCountTables := `
		SELECT COUNT(tablename)
			FROM pg_catalog.pg_tables
			WHERE tablename IN ('gauge', 'counter');
		`

		ctx, cancel = context.WithTimeout(ctxBase, time.Second)
		defer cancel()
		row := db.QueryRowContext(
			ctx,
			qCountTables,
		)
		var count int
		require.NoError(t, row.Scan(&count))
		assert.Zero(t, count)

		storage := NewPSQL(db)
		storage.Run(stopCtx, &wg)

		ctx, cancel = context.WithTimeout(ctxBase, time.Second)
		defer cancel()

		row = db.QueryRowContext(ctx, qCountTables)
		require.NoError(t, row.Scan(&count))
		assert.Equal(t, 2, count)
	})

	t.Run("Sequence update gauge by name", func(t *testing.T) {
		clearTables(t)
		storage := NewPSQL(db)
		storage.Run(stopCtx, &wg)
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
			assert.InDelta(t, v, actualValue, 0)
		}
	})

	t.Run("Sequence update counter by name", func(t *testing.T) {
		clearTables(t)
		storage := NewPSQL(db)
		storage.Run(stopCtx, &wg)
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

	t.Run("Batch update", func(t *testing.T) {
		t.Run("Gauge (no doubles)", func(t *testing.T) {
			clearTables(t)
			storage := NewPSQL(db)
			storage.Run(stopCtx, &wg)

			var m0 metrics.Metrics
			m0.ID = "0"
			m0.MType = metrics.MTypeGauge
			m0v := 5.0
			m0.Value = &m0v

			var m1 metrics.Metrics
			m1.ID = "1"
			m1.MType = metrics.MTypeGauge
			m1v := 7.0
			m1.Value = &m1v

			gaugeSlice := metrics.MetricsList{m0, m1}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := storage.UpdateGaugeList(ctx, gaugeSlice)
			require.NoError(t, err)

			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			rows, err := db.QueryContext(
				ctx,
				`SELECT name, value
				 FROM gauge 
				 ORDER BY name ASC;`,
			)
			require.NoError(t, err)
			defer rows.Close()

			var rowNumber int
			for rows.Next() {
				var (
					name  string
					value float64
				)
				require.NoError(t, rows.Scan(&name, &value))
				assert.Equal(t, gaugeSlice[rowNumber].ID, name)
				assert.InDelta(t, *gaugeSlice[rowNumber].Value, value, 0)
				rowNumber++
			}
			assert.NoError(t, rows.Err())
		})

		t.Run("Gauge (with doubles)", func(t *testing.T) {
			clearTables(t)
			storage := NewPSQL(db)
			storage.Run(stopCtx, &wg)

			var m0 metrics.Metrics
			m0.ID = "0"
			m0.MType = metrics.MTypeGauge
			m0v := 5.0
			m0.Value = &m0v

			var m1 metrics.Metrics
			m1.ID = "1"
			m1.MType = metrics.MTypeGauge
			m1v := 7.0
			m1.Value = &m1v

			var m2 metrics.Metrics
			m2.ID = "0"
			m2.MType = metrics.MTypeGauge
			m2v := 5.0
			m2.Value = &m2v

			gaugeSlice := metrics.MetricsList{m0, m1, m2}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := storage.UpdateGaugeList(ctx, gaugeSlice)
			require.NoError(t, err)

			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			rows, err := db.QueryContext(
				ctx,
				`SELECT name, value
				 FROM gauge 
				 ORDER BY name ASC;`,
			)
			require.NoError(t, err)
			defer rows.Close()

			expectedSlice := []metrics.Metrics{
				{ID: "0", MType: metrics.MTypeGauge, Value: gaugeSlice[2].Value},
				{ID: "1", MType: metrics.MTypeGauge, Value: gaugeSlice[1].Value},
			}

			var rowNumber int
			for rows.Next() {
				var (
					name  string
					value float64
				)
				require.NoError(t, rows.Scan(&name, &value))
				assert.Equal(t, expectedSlice[rowNumber].ID, name)
				assert.InDelta(t, *expectedSlice[rowNumber].Value, value, 0)
				rowNumber++
			}
			assert.NoError(t, rows.Err())
		})

		t.Run("Counter (no doubles)", func(t *testing.T) {
			clearTables(t)
			storage := NewPSQL(db)
			storage.Run(stopCtx, &wg)

			var m0 metrics.Metrics
			m0.ID = "0"
			m0.MType = metrics.MTypeCounter
			m0v := int64(5)
			m0.Delta = &m0v

			var m1 metrics.Metrics
			m1.ID = "1"
			m1.MType = metrics.MTypeCounter
			m1v := int64(7)
			m1.Delta = &m1v

			counterSlice := metrics.MetricsList{m0, m1}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := storage.UpdateCounterList(ctx, counterSlice)
			require.NoError(t, err)

			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			rows, err := db.QueryContext(
				ctx,
				`SELECT name, value
				 FROM counter 
				 ORDER BY name ASC;`,
			)
			require.NoError(t, err)
			defer rows.Close()

			var rowNumber int
			for rows.Next() {
				var (
					name  string
					delta int64
				)
				require.NoError(t, rows.Scan(&name, &delta))
				assert.Equal(t, counterSlice[rowNumber].ID, name)
				assert.Equal(t, *counterSlice[rowNumber].Delta, delta)
				rowNumber++
			}
			assert.NoError(t, rows.Err())
		})

		t.Run("Counter (with doubles)", func(t *testing.T) {
			clearTables(t)
			storage := NewPSQL(db)
			storage.Run(stopCtx, &wg)

			var m0 metrics.Metrics
			m0.ID = "0"
			m0.MType = metrics.MTypeCounter
			m0v := int64(5)
			m0.Delta = &m0v

			var m1 metrics.Metrics
			m1.ID = "1"
			m1.MType = metrics.MTypeCounter
			m1v := int64(7)
			m1.Delta = &m1v

			var m2 metrics.Metrics
			m2.ID = "0"
			m2.MType = metrics.MTypeCounter
			m2v := int64(10)
			m2.Delta = &m2v

			counterSlice := metrics.MetricsList{m0, m1, m2}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := storage.UpdateCounterList(ctx, counterSlice)
			require.NoError(t, err)

			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			rows, err := db.QueryContext(
				ctx,
				`SELECT name, value
				 FROM counter 
				 ORDER BY name ASC;`,
			)
			require.NoError(t, err)
			defer rows.Close()

			var expectM0 metrics.Metrics
			expectM0.ID = "0"
			expectM0.MType = metrics.MTypeCounter
			expectM0Delta := (*counterSlice[0].Delta) + (*counterSlice[2].Delta)
			expectM0.Delta = &expectM0Delta

			var expectM1 metrics.Metrics
			expectM1.ID = "1"
			expectM1.MType = metrics.MTypeCounter
			expectM1Delta := *counterSlice[1].Delta
			expectM1.Delta = &expectM1Delta

			expectedSlice := []metrics.Metrics{expectM0, expectM1}

			var rowNumber int
			for rows.Next() {
				var (
					name  string
					delta int64
				)
				require.NoError(t, rows.Scan(&name, &delta))
				assert.Equal(t, expectedSlice[rowNumber].ID, name)
				assert.Equal(t, *expectedSlice[rowNumber].Delta, delta)
				rowNumber++
			}
			assert.NoError(t, rows.Err())
		})

	})

	t.Run("Read counter by name", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run(stopCtx, &wg)
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
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			_, err := db.ExecContext(
				ctx,
				`INSERT INTO counter (name, value)
				 VALUES ($1, $2);`,
				metricName,
				expected,
			)
			require.NoError(t, err)
			ctx, cancel = context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			actualValue, err := storage.ReadCounterByName(
				ctx, metricName,
			)
			require.NoError(t, err)
			assert.Equal(t, expected, actualValue)
		})
	})

	t.Run("Read gauge by name", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run(stopCtx, &wg)
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
			assert.InDelta(t, expected, actualValue, 0)
		})

		t.Run("Should return entry value", func(t *testing.T) {
			clearTables(t)
			expected := float64(10)
			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			_, err := db.ExecContext(
				ctx,
				`INSERT INTO gauge (name, value)
				 VALUES ($1, $2);`,
				metricName,
				expected,
			)
			require.NoError(t, err)
			ctx, cancel = context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			actualValue, err := storage.ReadGaugeByName(
				ctx, metricName,
			)
			require.NoError(t, err)
			assert.InDelta(t, expected, actualValue, 0)
		})
	})

	t.Run("Read counter", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run(stopCtx, &wg)
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

			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			_, err := db.ExecContext(
				ctx,
				`INSERT INTO counter (name, value) VALUES `+
					strings.Join(insertValues, ", ")+";",
			)
			require.NoError(t, err)

			ctx, cancel = context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			counterData, err := storage.ReadCounter(ctx)
			require.NoError(t, err)
			assert.Len(t, counterData, expectedDataLen)
			for _, counter := range testList {
				assert.Equal(t, counter.value, counterData[counter.name])
			}
		})
	})

	t.Run("Read gauge", func(t *testing.T) {
		storage := NewPSQL(db)
		storage.Run(stopCtx, &wg)
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

			ctx, cancel := context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			_, err := db.ExecContext(
				ctx,
				`INSERT INTO gauge (name, value) VALUES `+
					strings.Join(insertValues, ", ")+";",
			)
			require.NoError(t, err)

			ctx, cancel = context.WithTimeout(ctxBase, time.Second)
			defer cancel()
			gaugeData, err := storage.ReadGauge(ctx)
			require.NoError(t, err)
			assert.Len(t, gaugeData, expectedDataLen)
			for _, gauge := range testList {
				assert.InDelta(t, gauge.value, gaugeData[gauge.name], 0)
			}
		})
	})
}
