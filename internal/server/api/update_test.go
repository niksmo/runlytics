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

type MockUpdateService struct {
	err bool
}

func (service *MockUpdateService) Update(metrics *schemas.Metrics) error {
	if service.err {
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

func TestUpdateByJSONHandler(t *testing.T) {

	newMetrics := func(id string, mType string, delta int64, value float64) *schemas.Metrics {
		return &schemas.Metrics{
			ID:    id,
			MType: mType,
			Delta: &delta,
			Value: &value,
		}
	}

	testRequest := func(
		t *testing.T,
		mux *chi.Mux,
		method string,
		path string,
		contentType string,
		body io.Reader,
	) (*http.Response, []byte) {
		ts := httptest.NewServer(mux)
		defer ts.Close()

		req, err := http.NewRequest(method, ts.URL+path, body)
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
		path        string
		want        want
		contentType string
		reqData     []byte
		service     *MockUpdateService
	}

	tests := []test{
		// not allowed methods
		{
			name:   "GET not allowed",
			method: http.MethodGet,
			path:   "/update/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "PUT not allowed",
			method: http.MethodPut,
			path:   "/update/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "PATCH not allowed",
			method: http.MethodPatch,
			path:   "/update/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "DELETE not allowed",
			method: http.MethodDelete,
			path:   "/update/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "HEAD not allowed",
			method: http.MethodHead,
			path:   "/update/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},
		{
			name:   "OPTIONS not allowed",
			method: http.MethodOptions,
			path:   "/update/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				resData:    nil,
			},
		},

		// POST
		{
			name:   "Should update metrics",
			method: http.MethodPost,
			path:   "/update/",
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
			reqData: []byte(`{
			    "id":"test",
		        "type":"test",
				"delta":321,
				"value":432.1
			}`),
			service: &MockUpdateService{err: false},
		},
		{
			name:   "Wrong Content-Type",
			method: http.MethodPost,
			path:   "/update/",
			want: want{
				statusCode: http.StatusUnsupportedMediaType,
				resData:    nil,
			},
			contentType: "text/plain",
			reqData: []byte(`{
			    "id":"test",
		        "type":"test",
				"delta":321,
				"value":432.1
			}`),
			service: &MockUpdateService{err: false},
		},
		{
			name:   "Wrong metrics scheme or bad JSON",
			method: http.MethodPost,
			path:   "/update/",
			want: want{
				statusCode: http.StatusBadRequest,
				resData:    nil,
			},
			contentType: "application/json",
			reqData: []byte(`{
			    "id":"test",
		        "type":"test",
				"delta":321.77897,
				"value":432.1
			}`),
			service: &MockUpdateService{err: false},
		},
		{
			name:   "Service should return error",
			method: http.MethodPost,
			path:   "/update/",
			want: want{
				statusCode: http.StatusBadRequest,
				resData:    nil,
			},
			contentType: "application/json",
			reqData: []byte(`{
			    "id":"test",
		        "type":"test",
				"delta":321,
				"value":432.1
			}`),
			service: &MockUpdateService{err: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := chi.NewRouter()
			SetUpdateHandler(mux, test.service)
			res, resBody := testRequest(t,
				mux,
				test.method,
				test.path,
				test.contentType,
				bytes.NewReader(test.reqData),
			)
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.resData != nil {
				require.Equal(t, JSONMediaType, res.Header.Get(ContentTypePath))

				var resData schemas.Metrics
				require.NoError(t, json.Unmarshal(resBody, &resData))
				assert.Equal(t, *test.want.resData, resData)
			}
		})
	}
}

func TestUpdateByURLParamsHandler(t *testing.T) {

	testRequest := func(
		t *testing.T,
		mux *chi.Mux,
		method string,
		path string,
		body io.Reader,
	) (*http.Response, []byte) {
		ts := httptest.NewServer(mux)
		defer ts.Close()

		req, err := http.NewRequest(method, ts.URL+path, body)
		require.NoError(t, err)

		res, err := ts.Client().Do(req)
		require.NoError(t, err)

		resData, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		return res, resData
	}

	type want struct {
		statusCode int
	}

	type test struct {
		name    string
		method  string
		path    string
		want    want
		service *MockUpdateService
	}

	tests := []test{
		// not allowed methods
		{
			name:   "GET not allowed",
			method: http.MethodGet,
			path:   "/update/gauge/testName/0",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "PUT not allowed",
			method: http.MethodPut,
			path:   "/update/gauge/testName/0",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "PATCH not allowed",
			method: http.MethodPatch,
			path:   "/update/gauge/testName/0",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "DELETE not allowed",
			method: http.MethodDelete,
			path:   "/update/gauge/testName/0",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "HEAD not allowed",
			method: http.MethodHead,
			path:   "/update/gauge/testName/0",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "OPTIONS not allowed",
			method: http.MethodOptions,
			path:   "/update/gauge/testName/0",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},

		// POST
		{
			name:   "Should update metrics",
			method: http.MethodPost,
			path:   "/update/gauge/testName/1234.56",
			want: want{
				statusCode: http.StatusOK,
			},
			service: &MockUpdateService{err: false},
		},
		{
			name:   "Wrong gauge format",
			method: http.MethodPost,
			path:   "/update/gauge/testName/null",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			service: &MockUpdateService{err: false},
		},
		{
			name:   "Wrong counter format",
			method: http.MethodPost,
			path:   "/update/counter/testName/0.2394871234",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			service: &MockUpdateService{err: false},
		},
		{
			name:   "Service should return error",
			method: http.MethodPost,
			path:   "/update/counter/testName/123456",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			service: &MockUpdateService{err: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := chi.NewRouter()
			SetUpdateHandler(mux, test.service)
			res, _ := testRequest(t,
				mux,
				test.method,
				test.path,
				http.NoBody,
			)
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)
		})
	}
}
