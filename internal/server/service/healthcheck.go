package service

import (
	"context"
	"errors"
	"strings"

	"github.com/niksmo/runlytics/pkg/di"
)

var (
	ErrStorage = errors.New("storage: down")
)

// HealthCheckService works with storage and provides Check method.
type HealthCheckService struct {
	storage di.IStorage
}

// NewHealthCheckService returns HealthCheckService pointer.
func NewHealthCheckService(storage di.IStorage) *HealthCheckService {
	return &HealthCheckService{storage: storage}
}

// Check sends test to storage and returns [ErrStorage] if occured.
func (s *HealthCheckService) Check(ctx context.Context) error {
	var errs errS
	if err := s.pingDB(ctx); err != nil {
		errs = append(errs, err)
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

func (s *HealthCheckService) pingDB(ctx context.Context) error {
	if err := s.storage.Ping(ctx); err != nil {
		return ErrStorage
	}
	return nil
}

// errS implements error interface.
type errS []error

// Error returns string that groups collected errors.
func (errs errS) Error() string {
	s := make([]string, 0, len(errs))
	for _, err := range errs {
		s = append(s, err.Error())
	}
	return strings.Join(s, "; ")
}

// Unwrap implements base error interface method.
func (errs errS) Unwrap() []error {
	if len(errs) == 0 {
		return nil
	}
	return errs
}
