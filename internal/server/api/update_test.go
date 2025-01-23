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

func newMetrics(id string, mType string, delta int64, value float64) *schemas.Metrics {
	return &schemas.Metrics{id, mType, &delta, &value}
}

type MockUpdateService struct{}

func (service *MockUpdateService) Update(metrics *schemas.Metrics) error {
	if metrics.ID == "error" {
		return errors.New("test error")
	}

	metrics.ID = "update"
	metrics.MType = "update"

	delta := int64(123)
	metrics.Delta = &delta

	value := 123.4
	metrics.Value = &value

	return nil
}

func TestUpdateHandler(t *testing.T) {

	testRequest := func(
		t *testing.T,
		mux *chi.Mux,
		method string,
		contentType string,
		body io.Reader,
	) (*http.Response, []byte) {
		ts := httptest.NewServer(mux)
		defer ts.Close()

		req, err := http.NewRequest(method, ts.URL+"/update/", body)
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
		reqData     map[string]any
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

		//post
		{
			name:   "Should update metrics",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				resData: newMetrics(
					"update",
					"update",
					123,
					123.4,
				),
			},
			contentType: "application/json",
			reqData: map[string]any{
				"id":    "test",
				"type":  "test",
				"delta": 321,
				"value": 432.1,
			},
		},
		{
			name:   "Wrong Content-Type",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusUnsupportedMediaType,
				resData:    nil,
			},
			contentType: "text/plain",
			reqData: map[string]any{
				"id":    "test",
				"type":  "test",
				"delta": 321,
				"value": 432.1,
			},
		},
		{
			name:   "Wrong metrics scheme or bad JSON",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				resData:    nil,
			},
			contentType: "application/json",
			reqData: map[string]any{
				"id":    "test",
				"type":  "test",
				"delta": 321.77897,
				"value": 432.1,
			},
		},
		{
			name:   "Service return error",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				resData:    nil,
			},
			contentType: "application/json",
			reqData: map[string]any{
				"id":    "error",
				"type":  "test",
				"delta": 321.77897,
				"value": 432.1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := chi.NewRouter()
			mockService := &MockUpdateService{}
			SetUpdateHandler(mux, mockService)
			reqBody, err := json.Marshal(test.reqData)
			require.NoError(t, err)

			res, resBody := testRequest(t,
				mux,
				test.method,
				test.contentType,
				bytes.NewReader(reqBody),
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
