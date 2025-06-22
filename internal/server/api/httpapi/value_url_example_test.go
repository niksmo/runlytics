package httpapi_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/api/httpapi"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/mock"
)

// ValueByJSONService represents mocked ValueService project component.
type ValueByURLService struct {
	mock.Mock
}

// Read assumes write "123.45" to metrics value.
func (service *ValueByURLService) Read(
	ctx context.Context, m *metrics.Metrics,
) error {
	retArgs := service.Called(context.Background(), m)
	if m != nil {
		m.Value = 123.45
	}
	return retArgs.Error(0)
}

func ExampleSetValueHandler_readByURLParams() {
	id := "0"
	mType := metrics.MTypeGauge

	var scheme metrics.Metrics
	scheme.ID = id
	scheme.MType = mType

	valueService := new(ValueByURLService)
	valueService.On("Read", context.Background(), &scheme).Return(nil)

	mux := chi.NewRouter()
	httpapi.SetValueHandler(mux, valueService)

	s := httptest.NewServer(mux)
	defer s.Close()

	reqURL, err := url.JoinPath(s.URL+"/value", mType, id)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		reqURL,
		http.NoBody,
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
	io.Copy(os.Stdout, res.Body)
	// Output:
	// 200
	// 123.45
}
