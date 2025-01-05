package router

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepoRead struct {
	getCounterCalls, getGaugeCalls int
}

func (fr *fakeRepoRead) GetCounter(name string) (int64, error) {
	fr.getCounterCalls++
	if name != "testName" {
		return 0, errors.New("metric not exists")
	}
	var ret int64 = 123
	return ret, nil
}

func (fr *fakeRepoRead) GetGauge(name string) (float64, error) {
	fr.getGaugeCalls++
	if name != "testName" {
		return 0, errors.New("metric not exists")
	}
	var ret float64 = 0.123
	return ret, nil
}

func TestValueHandler(t *testing.T) {

	testRequest := func(t *testing.T, mux *chi.Mux, method string, path string) (*http.Response, []byte) {
		ts := httptest.NewServer(mux)
		defer ts.Close()

		req, err := http.NewRequest(method, ts.URL+path, nil)
		require.NoError(t, err)

		res, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		return res, body
	}

	type want struct {
		statusCode int
		repoCalls  int
		body       string
	}

	type test struct {
		name       string
		method     string
		want       want
		path       string
		metricType server.MetricType
	}

	tests := []test{
		{
			name:   "POST not allowed",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/value/gauge/testName",
			metricType: gauge,
		},
		{
			name:   "PUT not allowed",
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/value/gauge/testName",
			metricType: gauge,
		},
		{
			name:   "PATCH not allowed",
			method: http.MethodPatch,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/value/gauge/testName",
			metricType: gauge,
		},
		{
			name:   "DELETE not allowed",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/value/gauge/testName",
			metricType: gauge,
		},
		{
			name:   "HEAD not allowed",
			method: http.MethodHead,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/value/gauge/testName",
			metricType: gauge,
		},
		{
			name:   "OPTIONS not allowed",
			method: http.MethodOptions,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				repoCalls:  0,
			},
			path:       "/value/gauge/testName",
			metricType: gauge,
		},
		{
			name:   "Should return gauge value",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
				body:       "0.123",
			},
			path:       "/value/gauge/testName",
			metricType: gauge,
		},
		{
			name:   "Should return counter value",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
				body:       "123",
			},
			path:       "/value/counter/testName",
			metricType: counter,
		},
		{
			name:   "Bad gauge metrics name",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusNotFound,
				repoCalls:  1,
			},
			path:       "/value/gauge/testName1",
			metricType: gauge,
		},
		{
			name:   "Bad counter metrics name",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusNotFound,
				repoCalls:  1,
			},
			path:       "/value/counter/testName1",
			metricType: counter,
		},
		{
			name:   "Bad gauge type",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusNotFound,
				repoCalls:  0,
			},
			path:       "/value/gauge1/testName",
			metricType: gauge,
		},
		{
			name:   "Bad counter type",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusNotFound,
				repoCalls:  0,
			},
			path:       "/value/counter1/testName",
			metricType: counter,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepoRead{}
			router := chi.NewRouter()
			SetValueRoute(router, repo)
			res, body := testRequest(t, router, test.method, test.path)

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.metricType == gauge {
				assert.Equal(t, test.want.repoCalls, repo.getGaugeCalls)
			}

			if test.metricType == counter {
				assert.Equal(t, test.want.repoCalls, repo.getCounterCalls)
			}

			if res.StatusCode == http.StatusOK {
				assert.Equal(t, test.want.body, string(body))
			}
		})
	}
}
