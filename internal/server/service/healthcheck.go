package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrDB = errors.New("database: down")
)

// HealthCheckService works with database and provides Check method.
type HealthCheckService struct {
	db *sql.DB
}

// NewHealthCheckService returns HealthCheckService pointer.
func NewHealthCheckService(db *sql.DB) *HealthCheckService {
	return &HealthCheckService{db}
}

// Check sends test to database and returns [ErrDB] if occured.
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
	if err := s.db.PingContext(ctx); err != nil {
		return ErrDB
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

// Unwrap implements base error interface.
func (errs errS) Unwrap() []error {
	if len(errs) == 0 {
		return nil
	}
	return errs
}
