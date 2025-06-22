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

// UpdateByURLService represents mocked UpdateService project component.
type UpdateByURLService struct {
	mock.Mock
}

func (service *UpdateByURLService) Update(
	ctx context.Context, m *metrics.Metrics,
) error {
	retArgs := service.Called(context.Background(), m)
	return retArgs.Error(0)
}

func ExampleSetUpdateHandler_updateByURLParams() {
	id := "0"
	mType := metrics.MTypeGauge
	value := "123.45"
	schemeValue := 123.45

	var scheme metrics.Metrics
	scheme.ID = id
	scheme.MType = mType
	scheme.Value = schemeValue

	updateService := new(UpdateByURLService)
	updateService.On(
		"Update", context.Background(), &scheme,
	).Return(nil)

	mux := chi.NewRouter()
	httpapi.SetUpdateHandler(mux, updateService)

	s := httptest.NewServer(mux)
	defer s.Close()

	reqURL, err := url.JoinPath(s.URL+"/update", mType, id, value)
	if err != nil {
		fmt.Println(err)
		return
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
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
