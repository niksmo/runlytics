package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type CheckErr []error

func (errSlice CheckErr) Error() string {
	s := make([]string, 0, len(errSlice))
	for _, err := range errSlice {
		s = append(s, err.Error())
	}
	return strings.Join(s, "; ")
}

func (errSlice CheckErr) Unwrap() []error {
	if len(errSlice) == 0 {
		return nil
	}
	return errSlice
}

type HealthCheckService struct {
	db *sql.DB
}

func NewHealthCheckService(db *sql.DB) *HealthCheckService {
	return &HealthCheckService{db}
}

func (service *HealthCheckService) Check(ctx context.Context) error {
	var errList CheckErr
	if err := service.pingDB(ctx); err != nil {
		errList = append(errList, fmt.Errorf("database: %w", err))
	}
	if len(errList) != 0 {
		return errList
	}
	return nil
}

func (service *HealthCheckService) pingDB(ctx context.Context) error {
	if err := service.db.PingContext(ctx); err != nil {
		return errors.New("down")
	}
	return nil
}
