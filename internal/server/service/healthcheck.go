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

func (service *HealthCheckService) Check() error {
	return service.checkDataBase()
}

func (service *HealthCheckService) checkDataBase() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return service.db.PingContext(ctx)
}
