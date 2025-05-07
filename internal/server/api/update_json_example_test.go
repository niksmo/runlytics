package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/pkg/httpserver/header"
	"github.com/niksmo/runlytics/pkg/httpserver/mime"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/mock"
)

// UpdateByJSONService represents mocked UpdateService project component.
type UpdateByJSONService struct {
	mock.Mock
}

func (service *UpdateByJSONService) Update(
	ctx context.Context, m *metrics.Metrics,
) error {
	retArgs := service.Called(context.Background(), m)
	return retArgs.Error(0)
}

func ExampleSetUpdateHandler_updateByJSON() {
	var scheme metrics.Metrics
	scheme.ID = "0"
	scheme.MType = metrics.MTypeGauge
	value := 123.45
	scheme.Value = &value

	updateService := new(UpdateByJSONService)
	updateService.On("Update", context.Background(), &scheme).Return(nil)

	mux := chi.NewRouter()
	api.SetUpdateHandler(mux, updateService)

	s := httptest.NewServer(mux)
	defer s.Close()

	var reqData bytes.Buffer
	if err := json.NewEncoder(&reqData).Encode(scheme); err != nil {
		fmt.Println(err)
		return
	}

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, s.URL+"/update/", &reqData,
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
	io.Copy(os.Stdout, res.Body)

	// Output:
	// 200
	// {"value":123.45,"id":"0","type":"gauge"}
}
