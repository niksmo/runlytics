package storage

import (
	"context"
	"sync"
)

type Storage interface {
	UpdateCounterByName(ctx context.Context, name string, value int64) (int64, error)
	UpdateGaugeByName(ctx context.Context, name string, value float64) (float64, error)
	ReadCounterByName(ctx context.Context, name string) (int64, error)
	ReadGaugeByName(ctx context.Context, name string) (float64, error)
	ReadGauge(context.Context) (map[string]float64, error)
	ReadCounter(context.Context) (map[string]int64, error)
	CheckDB(context.Context) error
	Run(context.Context, *sync.WaitGroup)
}
