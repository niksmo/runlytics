package api_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/pkg/httpserver/header"
	"github.com/niksmo/runlytics/pkg/httpserver/mime"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUpdateService struct {
	mock.Mock
}

func (service *MockUpdateService) Update(
	ctx context.Context, m *metrics.Metrics,
) error {
	retArgs := service.Called(context.Background(), m)
	return retArgs.Error(0)
}

func TestUpdateByJSONHandler(t *testing.T) {
	makeURL := func(serverURL string) string {
		return serverURL + "/update/"
	}

	t.Run("Not allowed methods", func(t *testing.T) {
		methods := []string{
			http.MethodGet,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		}
		mockService := new(MockUpdateService)
		mockService.On("Update", context.Background(), nil).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{"id": "0", "type": "gauge", "value": 123.45}`,
			)

			req, err := http.NewRequestWithContext(
				context.Background(), method, makeURL(s.URL), reqBody,
			)
			require.NoError(t, err)
			req.Header.Set(header.ContentType, mime.JSON)

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Len(t, data, 0)

			mockService.AssertNumberOfCalls(t, "Update", 0)
		}
	})

	t.Run("Not allowed Content-Type", func(t *testing.T) {
		mockService := new(MockUpdateService)
		mockService.On("Update", context.Background(), nil).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{"id": "0", "type": "gauge", "value": 123.45}`,
		)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.TEXT)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)
		assert.Len(t, data, 0)

		mockService.AssertNumberOfCalls(t, "Update", 0)
	})

	t.Run("Bad JSON", func(t *testing.T) {
		mockService := new(MockUpdateService)
		mockService.On("Update", context.Background(), nil).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{"id": "0", "type": "gauge, "value": 123.45,}`,
		)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		mockService.AssertNumberOfCalls(t, "Update", 0)
	})

	t.Run("Invalid metrics payload", func(t *testing.T) {
		var schemeReq metrics.Metrics
		schemeReq.ID = ""
		schemeReq.MType = metrics.MTypeGauge

		mockService := new(MockUpdateService)
		mockService.On("Update", context.Background(), &schemeReq).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		res.Body.Close()

		mockService.AssertNumberOfCalls(t, "Update", 0)
	})

	t.Run("Error on Update", func(t *testing.T) {
		var schemeReq metrics.Metrics
		schemeReq.ID = "0"
		schemeReq.MType = metrics.MTypeGauge
		value := 123.45
		schemeReq.Value = &value
		updateErr := errors.New("test error")

		mockService := new(MockUpdateService)
		mockService.On(
			"Update", context.Background(), &schemeReq,
		).Return(updateErr)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, server.ErrInternal.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		var schemeReq metrics.Metrics
		schemeReq.ID = "0"
		schemeReq.MType = metrics.MTypeGauge
		value := 123.45
		schemeReq.Value = &value

		mockService := new(MockUpdateService)
		mockService.On("Update", context.Background(), &schemeReq).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		var bufReq bytes.Buffer
		err := json.NewEncoder(&bufReq).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), &bufReq,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		expect := `{"id":"0","type":"gauge","value":123.45}`
		assert.JSONEq(t, expect, string(data))

		mockService.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("Encoding", func(t *testing.T) {
		t.Run("Allow only gzip", func(t *testing.T) {
			mockService := new(MockUpdateService)
			mockService.On("Update", context.Background(), nil).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)
			api.SetUpdateHandler(mux, mockService)

			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{"id":"0","type":"gauge","value":123.45}`,
			)

			req, err := http.NewRequestWithContext(
				context.Background(), http.MethodPost, makeURL(s.URL), reqBody,
			)
			require.NoError(t, err)
			req.Header.Set(header.ContentType, mime.JSON)
			req.Header.Set(header.ContentEncoding, "br")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Len(t, data, 0)

			mockService.AssertNumberOfCalls(t, "Update", 0)
		})

		t.Run("Send gzip, accept non-compressed", func(t *testing.T) {
			var schemeReq metrics.Metrics
			schemeReq.ID = "0"
			schemeReq.MType = metrics.MTypeGauge
			value := 123.45
			schemeReq.Value = &value

			mockService := new(MockUpdateService)
			mockService.On(
				"Update", context.Background(), &schemeReq,
			).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)
			api.SetUpdateHandler(mux, mockService)

			s := httptest.NewServer(mux)
			defer s.Close()

			var buf bytes.Buffer
			gzipWriter := gzip.NewWriter(&buf)
			err := json.NewEncoder(gzipWriter).Encode(&schemeReq)
			require.NoError(t, err)
			err = gzipWriter.Close()
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				context.Background(), http.MethodPost, makeURL(s.URL), &buf,
			)
			require.NoError(t, err)
			req.Header.Set(header.ContentType, mime.JSON)
			req.Header.Set(header.ContentEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			assert.Zero(t, res.Header.Get(header.ContentEncoding))

			data, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			res.Body.Close()
			expect := `{"id":"0","type":"gauge","value":123.45}`
			assert.JSONEq(t, expect, string(data))

			mockService.AssertNumberOfCalls(t, "Update", 1)
		})

		t.Run("Send non-compressed, accept gzip", func(t *testing.T) {
			var schemeReq metrics.Metrics
			schemeReq.ID = "0"
			schemeReq.MType = metrics.MTypeGauge
			value := 123.45
			schemeReq.Value = &value

			mockService := new(MockUpdateService)
			mockService.On(
				"Update", context.Background(), &schemeReq,
			).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)
			api.SetUpdateHandler(mux, mockService)

			s := httptest.NewServer(mux)
			defer s.Close()

			var bufReq bytes.Buffer
			err := json.NewEncoder(&bufReq).Encode(&schemeReq)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				context.Background(), http.MethodPost, makeURL(s.URL), &bufReq,
			)
			require.NoError(t, err)
			req.Header.Set(header.ContentType, mime.JSON)
			req.Header.Set(header.AcceptEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			gzipReader, err := gzip.NewReader(res.Body)
			require.NoError(t, err)
			data, err := io.ReadAll(gzipReader)
			require.NoError(t, err)
			gzipReader.Close()
			res.Body.Close()
			expect := `{"id":"0","type":"gauge","value":123.45}`
			assert.JSONEq(t, expect, string(data))

			mockService.AssertNumberOfCalls(t, "Update", 1)
		})
	})
}

func TestUpdateByURLParamsHandler(t *testing.T) {
	makeURL := func(serverURL, mType, mName, mValue string) string {
		testURL, _ := url.JoinPath(serverURL+"/update", mType, mName, mValue)
		return testURL
	}
	t.Run("Not allowed methods", func(t *testing.T) {
		methods := []string{
			http.MethodGet,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		}
		mockService := new(MockUpdateService)
		mockService.On("Update", context.Background(), nil).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			req, err := http.NewRequestWithContext(
				context.Background(),
				method,
				makeURL(s.URL, "gauge", "Alloc", "123.45"),
				http.NoBody,
			)
			require.NoError(t, err)

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Len(t, data, 0)

			mockService.AssertNumberOfCalls(t, "Update", 0)
		}
	})

	t.Run("Invalid metrics payload", func(t *testing.T) {
		id := "0"
		mType := "bgauge"
		value := "123.45"

		mockService := new(MockUpdateService)
		mockService.On(
			"Update", context.Background(), nil,
		).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			makeURL(s.URL, mType, id, value),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		res.Body.Close()
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		mockService.AssertNumberOfCalls(t, "Update", 0)
	})

	t.Run("Error on update", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeGauge
		valueStr := "123.45"

		var scheme metrics.Metrics
		scheme.ID = id
		scheme.MType = mType
		schemeValue := 123.45
		scheme.Value = &schemeValue

		updateErr := errors.New("test error")

		mockService := new(MockUpdateService)
		mockService.On(
			"Update", context.Background(), &scheme,
		).Return(updateErr)

		mux := chi.NewRouter()

		api.SetUpdateHandler(mux, mockService)
		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			makeURL(s.URL, mType, id, valueStr),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, server.ErrInternal.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeGauge
		expectedValue := "123.45"

		var schemeReq metrics.Metrics
		schemeReq.ID = id
		schemeReq.MType = mType
		schemeValue := 123.45
		schemeReq.Value = &schemeValue

		mockService := new(MockUpdateService)
		mockService.On(
			"Update", context.Background(), &schemeReq,
		).Return(nil)

		mux := chi.NewRouter()
		api.SetUpdateHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			makeURL(s.URL, mType, id, expectedValue),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, expectedValue, string(data))

		mockService.AssertNumberOfCalls(t, "Update", 1)
	})
}
