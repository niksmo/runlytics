package service

import (
	"context"
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

type DBChecker interface {
	CheckDB(context.Context) error
}

type HealthCheckService struct {
	dbChecker DBChecker
}

func NewHealthCheckService(dbChecker DBChecker) *HealthCheckService {
	return &HealthCheckService{dbChecker}
}

func (service *HealthCheckService) Check(ctx context.Context) error {
	var errList CheckErr
	if err := service.dbChecker.CheckDB(ctx); err != nil {
		errList = append(errList, err)
	}
	if len(errList) != 0 {
		return errList
	}
	return nil
}
