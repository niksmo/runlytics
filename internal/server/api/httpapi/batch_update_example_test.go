package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/api/httpapi"
	"github.com/niksmo/runlytics/internal/server/app/http/header"
	"github.com/niksmo/runlytics/internal/server/app/http/mime"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/mock"
)

// ExampleBatchUpdateService represents mocked BatchUpdateService project component.
type ExampleBatchUpdateService struct {
	mock.Mock
}

func (s *ExampleBatchUpdateService) BatchUpdate(
	ctx context.Context, ml metrics.MetricsList,
) error {
	retArgs := s.Called(context.Background(), ml)
	return retArgs.Error(0)
}

func ExampleSetBatchUpdateHandler() {
	m0value := 123.45
	m1delta := int64(12345)

	metricsList := metrics.MetricsList{
		{ID: "0", MType: metrics.MTypeGauge, Value: &m0value},
		{ID: "1", MType: metrics.MTypeCounter, Delta: &m1delta},
	}

	batchUpdateService := new(ExampleBatchUpdateService)
	batchUpdateService.On(
		"BatchUpdate", context.Background(), metricsList,
	).Return(nil)

	mux := chi.NewRouter()
	httpapi.SetBatchUpdateHandler(mux, batchUpdateService)

	s := httptest.NewServer(mux)
	defer s.Close()

	var reqData bytes.Buffer
	if err := json.NewEncoder(&reqData).Encode(metricsList); err != nil {
		fmt.Println(err)
		return
	}

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, s.URL+"/updates/", &reqData,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set(header.ContentType, mime.JSON)

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
