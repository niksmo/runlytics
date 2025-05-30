package httpapi_test

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
	"github.com/niksmo/runlytics/internal/server/api/httpapi"
	"github.com/niksmo/runlytics/internal/server/app/http/header"
	"github.com/niksmo/runlytics/internal/server/app/http/mime"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/mock"
)

// ValueByJSONService represents mocked ValueService project component.
type ValueByJSONService struct {
	mock.Mock
}

// Read assumes write "123.45" to metrics value.
func (service *ValueByJSONService) Read(
	ctx context.Context, m *metrics.Metrics,
) error {
	retArgs := service.Called(context.Background(), m)
	if m != nil {
		v := 123.45
		m.Value = &v
	}
	return retArgs.Error(0)
}

func ExampleSetValueHandler_readByJSON() {
	var scheme metrics.Metrics
	scheme.ID = "0"
	scheme.MType = metrics.MTypeGauge

	valueService := new(ValueByJSONService)
	valueService.On("Read", context.Background(), &scheme).Return(nil)

	mux := chi.NewRouter()
	httpapi.SetValueHandler(mux, valueService)

	s := httptest.NewServer(mux)
	defer s.Close()

	var reqData bytes.Buffer
	if err := json.NewEncoder(&reqData).Encode(scheme); err != nil {
		fmt.Println(err)
		return
	}

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, s.URL+"/value/", &reqData,
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
