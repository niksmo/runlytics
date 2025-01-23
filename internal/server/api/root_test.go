package api

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockHTMLService struct {
	err bool
}

func (service *MockHTMLService) RenderMetricsList(buf *bytes.Buffer) error {
	if service.err {
		return errors.New("test error")
	}

	buf.WriteString("test response data")
	return nil
}

func TestHTMLHandler(t *testing.T) {
	testRequest := func(
		t *testing.T,
		mux *chi.Mux,
		method string,
		body io.Reader,
	) (*http.Response, []byte) {
		ts := httptest.NewServer(mux)
		defer ts.Close()

		req, err := http.NewRequest(method, ts.URL+"/", body)
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
		name    string
		method  string
		want    want
		service *MockHTMLService
	}

	tests := []test{
		// not allowed methods
		{
			name:   "POST not allowed",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "PUT not allowed",
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "PATCH not allowed",
			method: http.MethodPatch,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "DELETE not allowed",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "HEAD not allowed",
			method: http.MethodHead,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "OPTIONS not allowed",
			method: http.MethodOptions,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},

		//GET
		{
			name:   "Should response OK",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				resData:    "test response data",
			},
			service: &MockHTMLService{err: false},
		},
		{
			name:   "Service should return error",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			service: &MockHTMLService{err: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := chi.NewRouter()
			SetHTMLHandler(mux, test.service)
			res, resBody := testRequest(t, mux, test.method, nil)
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.resData != "" {
				assert.Equal(t, test.want.resData, string(resBody))
			}
		})
	}
}
