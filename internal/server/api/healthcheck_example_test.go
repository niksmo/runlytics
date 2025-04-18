package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/stretchr/testify/mock"
)

// ExampleHealthCheckService represents mocked HealthCheckService project component.
type ExampleHealthCheckService struct {
	mock.Mock
}

func (s *ExampleHealthCheckService) Check(ctx context.Context) error {
	retArgs := s.Called(context.Background())
	return retArgs.Error(0)
}

func ExampleSetHealthCheckHandler() {
	healthCheckService := new(ExampleHealthCheckService)
	healthCheckService.On("Check", context.Background()).Return(nil)

	mux := chi.NewRouter()
	api.SetHealthCheckHandler(mux, healthCheckService)

	s := httptest.NewServer(mux)
	defer s.Close()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, s.URL+"/ping", http.NoBody,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := s.Client().Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	// Output:
	// 200
}
