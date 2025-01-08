package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepoMain struct {
	getDataCalls int
	gauge        map[string]float64
	counter      map[string]int64
}

func (fr *fakeRepoMain) GetData() (map[string]float64, map[string]int64) {
	fr.getDataCalls++
	return fr.gauge, fr.counter
}

func newFakeRepoMain(hasMetrics bool) *fakeRepoMain {
	r := &fakeRepoMain{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}

	if hasMetrics {
		r.gauge["gauge"] = 0.123
		r.counter["counter"] = 456
	}

	return r
}

func TestMainHandelr(t *testing.T) {

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
		contains   []string
	}

	type test struct {
		name       string
		method     string
		want       want
		path       string
		hasMetrics bool
	}

	tests := []test{
		{
			name:   "PUT not allowed",
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			path:       "/",
			hasMetrics: false,
		},
		{
			name:   "PATCH not allowed",
			method: http.MethodPatch,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			path:       "/",
			hasMetrics: false,
		},
		{
			name:   "POST not allowed",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			path:       "/",
			hasMetrics: false,
		},
		{
			name:   "DELETE not allowed",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			path:       "/",
			hasMetrics: false,
		},
		{
			name:   "HEAD not allowed",
			method: http.MethodHead,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			path:       "/",
			hasMetrics: false,
		},
		{
			name:   "OPTIONS not allowed",
			method: http.MethodOptions,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			path:       "/",
			hasMetrics: false,
		},
		{
			name:   "Should render metrics",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
				contains:   []string{"gauge: 0.123", "counter: 456"},
			},
			path:       "/",
			hasMetrics: true,
		},
		{
			name:   "Should render empty info",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
				contains:   []string{"metrics data is empty"},
			},
			path:       "/",
			hasMetrics: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeRepoMain(test.hasMetrics)
			router := chi.NewRouter()
			SetMainRoute(router, repo)

			res, body := testRequest(t, router, test.method, test.path)

			require.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.statusCode != http.StatusOK {
				return
			}

			require.Equal(t, "text/html; charset=utf-8", res.Header.Get("Content-Type"))

			data := string(body)

			for _, text := range test.want.contains {
				assert.True(t, strings.Contains(data, text))
			}

		})
	}
}
