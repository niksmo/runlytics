package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockValueService struct {
	err bool
}

func (service *MockValueService) Read(metrics *schemas.Metrics) error {
	if service.err {
		return errors.New("test error")

	}

	metrics.ID = "read"
	metrics.MType = "read"

	value := 123.4
	metrics.Value = &value

	return nil
}

func TestValueHandler(t *testing.T) {
	newMetrics := func(id string, mType string, value float64) *schemas.Metrics {
		return &schemas.Metrics{
			ID:    id,
			MType: mType,
			Value: &value,
		}
	}

	testRequest := func(
		t *testing.T,
		mux *chi.Mux,
		method string,
		contentType string,
		body io.Reader,
	) (*http.Response, []byte) {
		ts := httptest.NewServer(mux)
		defer ts.Close()

		req, err := http.NewRequest(method, ts.URL+"/value/", body)
		require.NoError(t, err)

		req.Header.Set("Content-Type", contentType)
		res, err := ts.Client().Do(req)
		require.NoError(t, err)

		resData, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		return res, resData
	}

	type want struct {
		statusCode int
		resData    *schemas.Metrics
	}

	type test struct {
		name        string
		method      string
		want        want
		contentType string
		reqData     []byte
		service     *MockValueService
	}

	tests := []test{
		// not allowed methods
		{
			name:   "GET not allowed",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "PUT not allowed",
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "PATCH not allowed",
			method: http.MethodPatch,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "DELETE not allowed",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "HEAD not allowed",
			method: http.MethodHead,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "OPTIONS not allowed",
			method: http.MethodOptions,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},

		//POST
		{
			name:   "Should read metrics",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				resData: newMetrics(
					"read",
					"read",
					123.4,
				),
			},
			contentType: "application/json",
			reqData:     []byte(`{"id": "test", "type": "test"}`),
			service:     &MockValueService{err: false},
		},
		{
			name:   "Wrong Content-Type",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusUnsupportedMediaType,
				resData:    nil,
			},
			contentType: "text/plain",
			reqData:     []byte(`{"id": "test", "type": "test"}`),
			service:     &MockValueService{err: false},
		},
		{
			name:   "Wrong metrics scheme or bad JSON",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				resData:    nil,
			},
			contentType: "application/json",
			reqData:     []byte(`{"id": "test", "type": "test}`),
			service:     &MockValueService{err: false},
		},
		{
			name:   "Service should return error",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
				resData:    nil,
			},
			contentType: "application/json",
			reqData:     []byte(`{"id": "test", "type": "test"}`),
			service:     &MockValueService{err: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := chi.NewRouter()
			SetReadHandler(mux, test.service)
			res, resBody := testRequest(t,
				mux,
				test.method,
				test.contentType,
				bytes.NewReader(test.reqData),
			)
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.resData != nil {
				var resData schemas.Metrics
				require.NoError(t, json.Unmarshal(resBody, &resData))
				assert.Equal(t, *test.want.resData, resData)
			}
		})
	}
}
