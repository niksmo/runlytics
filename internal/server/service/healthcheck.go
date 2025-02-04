package service

import (
	"context"
	"database/sql"
	"time"
)

type HealthCheckService struct {
	db *sql.DB
}

func NewHealthCheckService(db *sql.DB) *HealthCheckService {
	return &HealthCheckService{db}
}

func (service *HealthCheckService) Check(ctx context.Context) error {
	return service.checkDataBase(ctx)
}

func (service *HealthCheckService) checkDataBase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return service.db.PingContext(ctx)
}
