package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateHandler(t *testing.T) {

	testRequest := func(
		t *testing.T,
		mux *chi.Mux,
		method string,
		body io.Reader,
	) (*http.Response, []byte) {
		ts := httptest.NewServer(mux)
		defer ts.Close()

		req, err := http.NewRequest(method, ts.URL+"/update/", body)
		require.NoError(t, err)

		res, err := ts.Client().Do(req)
		require.NoError(t, err)

		resData, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		return res, resData
	}

	type want struct {
		statusCode int
		resData    string
	}

	type test struct {
		name   string
		method string
		want   want
		body   map[string]any
	}

	tests := []test{
		{
			name:   "GET not allowed",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/update/gauge/testName/0",
			metricType: gauge,
		},
		{
			name:   "PUT not allowed",
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/update/gauge/testName/0",
			metricType: gauge,
		},
		{
			name:   "PATCH not allowed",
			method: http.MethodPatch,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/update/gauge/testName/0",
			metricType: gauge,
		},
		{
			name:   "DELETE not allowed",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/update/gauge/testName/0",
			metricType: gauge,
		},
		{
			name:   "HEAD not allowed",
			method: http.MethodHead,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/update/gauge/testName/0",
			metricType: gauge,
		},
		{
			name:   "OPTIONS not allowed",
			method: http.MethodOptions,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/update/gauge/testName/0",
			metricType: gauge,
		},
		{
			name:   "Zero gauge",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			path:       "/update/gauge/testName/0",
			metricType: gauge,
		},
		{
			name:   "Positive gauge",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			path:       "/update/gauge/testName/0.412934812374",
			metricType: gauge,
		},
		{
			name:   "Negative gauge",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			path:       "/update/gauge/testName/-0.412934812374",
			metricType: gauge,
		},
		{
			name:   "Zero counter",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			path:       "/update/counter/testName/0",
			metricType: counter,
		},
		{
			name:   "Positive counter",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			path:       "/update/counter/testName/324567",
			metricType: counter,
		},
		{
			name:   "Negative counter",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			path:       "/update/counter/testName/-1234",
			metricType: counter,
		},
		{
			name:   "Wrong gauge path",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				repoCalls:  0,
			},
			path:       "/update/gaugee/testName/0.23234",
			metricType: gauge,
		},
		{
			name:   "Float value for counter metric",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				repoCalls:  0,
			},
			path:       "/update/counter/testName/0.2394871234",
			metricType: counter,
		},
		{
			name:   "Wrong counter path",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				repoCalls:  0,
			},
			path:       "/update/counters/testName/523",
			metricType: counter,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepoUpdate{}
			router := chi.NewRouter()
			SetUpdateRoute(router, repo)
			res, _ := testRequest(t, router, test.method, test.path)
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.metricType == gauge {
				assert.Equal(t, test.want.repoCalls, repo.setGaugeCalls)
			}

			if test.metricType == counter {
				assert.Equal(t, test.want.repoCalls, repo.addCounterCalls)
			}
		})
	}
}
