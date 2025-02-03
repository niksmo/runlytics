package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/metrics"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockUpdateService struct {
	err bool
}

func (service *MockUpdateService) Update(mData *metrics.Metrics) error {
	if service.err {
		return errors.New("test error")

	}

	mData.ID = "update"
	mData.MType = "update"

	delta := int64(123)
	mData.Delta = &delta

	value := 123.4
	mData.Value = &value

	return nil
}

func TestUpdateByJSONHandler(t *testing.T) {

	newMetrics := func(id string, mType string, delta int64, value float64) *metrics.Metrics {
		return &metrics.Metrics{
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
		resData    *metrics.Metrics
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
				require.Equal(t, JSONMediaType, res.Header.Get(ContentType))

				var resData metrics.Metrics
				require.NoError(t, json.Unmarshal(resBody, &resData))
				assert.Equal(t, *test.want.resData, resData)
			}
		})
	}
}

func TestUpdateByJSONHandlerGzip(t *testing.T) {
	mux := chi.NewRouter()
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)
	SetUpdateHandler(mux, &MockUpdateService{err: false})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	path := "/update/"

	requestBody := `{
	    "id":"test",
		"type":"test",
		"delta":321,
		"value":432.1
	}`

	successBody := `{
	    "id":"update",
        "type":"update",
        "delta":123,
        "value":123.4
	}`

	t.Run("send gzip", func(t *testing.T) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, err := gw.Write([]byte(requestBody))
		require.NoError(t, err)
		err = gw.Close()
		require.NoError(t, err)

		request, err := http.NewRequest("POST", srv.URL+path, &buf)
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "gzip")
		request.Header.Set("Accept-Encoding", "")
		res, err := srv.Client().Do(request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.JSONEq(t, successBody, string(data))
	})

	t.Run("accept gzip", func(t *testing.T) {
		request, err := http.NewRequest("POST", srv.URL+path, strings.NewReader(requestBody))
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept-Encoding", "gzip")
		res, err := srv.Client().Do(request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
		gr, err := gzip.NewReader(res.Body)
		defer res.Body.Close()
		require.NoError(t, err)
		data, err := io.ReadAll(gr)
		require.NoError(t, err)
		assert.JSONEq(t, successBody, string(data))
	})
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
