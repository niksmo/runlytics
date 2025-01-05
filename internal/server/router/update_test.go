package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/stretchr/testify/assert"
)

type fakeRepo struct {
	addCounterCalls, addGaugeCalls int
}

func (fr *fakeRepo) AddCounter(name string, value int64) {
	fr.addCounterCalls++
}
func (fr *fakeRepo) AddGauge(name string, value float64) {
	fr.addGaugeCalls++
}

func TestUpdateHandler(t *testing.T) {
	type want struct {
		statusCode int
		repoCalls  int
	}

	type test struct {
		name       string
		method     string
		want       want
		pathBase   string
		pathType   string
		pathName   string
		pathValue  string
		metricType server.MetricType
	}

	tests := []test{
		{
			name:   "Zero gauge",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			pathBase:   "/update",
			pathType:   "gauge",
			pathName:   "testName",
			pathValue:  "0",
			metricType: gauge,
		},
		{
			name:   "Positive gauge",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			pathBase:   "/update",
			pathType:   "gauge",
			pathName:   "testName",
			pathValue:  "0.412934812374",
			metricType: gauge,
		},
		{
			name:   "Negative gauge",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			pathBase:   "/update",
			pathType:   "gauge",
			pathName:   "testName",
			pathValue:  "-0.412934812374",
			metricType: gauge,
		},
		{
			name:   "Zero counter",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			pathBase:   "/update",
			pathType:   "counter",
			pathName:   "testName",
			pathValue:  "0",
			metricType: counter,
		},
		{
			name:   "Positive counter",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			pathBase:   "/update",
			pathType:   "counter",
			pathName:   "testName",
			pathValue:  "324567",
			metricType: counter,
		},
		{
			name:   "Negative counter",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
				repoCalls:  1,
			},
			pathBase:   "/update",
			pathType:   "counter",
			pathName:   "testName",
			pathValue:  "-1234",
			metricType: counter,
		},
		{
			name:   "Wrong gauge path",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				repoCalls:  0,
			},
			pathBase:   "/update",
			pathType:   "gaugee",
			pathName:   "testName",
			pathValue:  "0.23234",
			metricType: gauge,
		},
		{
			name:   "Float value for counter metric",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				repoCalls:  0,
			},
			pathBase:   "/update",
			pathType:   "counter",
			pathName:   "testName",
			pathValue:  "0.2394871234",
			metricType: counter,
		},
		{
			name:   "Wrong counter path",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				repoCalls:  0,
			},
			pathBase:   "/update",
			pathType:   "counters",
			pathName:   "testName",
			pathValue:  "523",
			metricType: counter,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepo{}
			metricHandler := &updateHandler{repo}
			updateHandler := metricHandler.update()
			url := strings.Join([]string{test.pathBase, test.pathType, test.pathName, test.pathValue}, "/")
			req := httptest.NewRequest(test.method, url, nil)
			req.SetPathValue("type", test.pathType)
			req.SetPathValue("name", test.pathName)
			req.SetPathValue("value", test.pathValue)
			w := httptest.NewRecorder()

			updateHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.metricType == gauge {
				assert.Equal(t, test.want.repoCalls, repo.addGaugeCalls)
			}

			if test.metricType == counter {
				assert.Equal(t, test.want.repoCalls, repo.addCounterCalls)
			}
		})
	}
}
