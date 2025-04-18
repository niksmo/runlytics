package api_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHTMLService struct {
	err  error
	data string
}

func (service *mockHTMLService) RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error {
	if service.err != nil {
		return service.err
	}

	buf.WriteString(service.data)
	return nil
}

func TestHTMLHandler(t *testing.T) {
	makeURL := func(serverURL string) string {
		return serverURL + "/"
	}

	t.Run("Not allowed methods", func(t *testing.T) {
		methods := []string{
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		}
		mux := chi.NewRouter()
		api.SetHTMLHandler(mux, nil)

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			req, err := http.NewRequest(method, makeURL(s.URL), http.NoBody)
			require.NoError(t, err)

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			require.NoError(t, err)
			assert.Len(t, data, 0)
		}
	})

	t.Run("Should return data", func(t *testing.T) {
		expectedData := "test"
		mux := chi.NewRouter()
		api.SetHTMLHandler(mux, &mockHTMLService{err: nil, data: expectedData})
		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequest(http.MethodGet, makeURL(s.URL), http.NoBody)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, expectedData, string(data))
	})

	t.Run("Should return internal error", func(t *testing.T) {
		mux := chi.NewRouter()
		api.SetHTMLHandler(mux, &mockHTMLService{err: errs.ErrInternal, data: ""})
		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequest(http.MethodGet, makeURL(s.URL), http.NoBody)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, res.StatusCode)

		rawData, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		data := strings.TrimSpace(string(rawData))
		assert.Equal(t, errs.ErrInternal.Error(), data)
	})
}
