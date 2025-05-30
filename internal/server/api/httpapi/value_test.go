package httpapi_test

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
	"github.com/niksmo/runlytics/internal/server/api/httpapi"
	"github.com/niksmo/runlytics/internal/server/app/http/middleware"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockValueService struct {
	mock.Mock
}

func (service *MockValueService) Read(
	ctx context.Context, m *metrics.Metrics,
) error {
	retArgs := service.Called(context.Background(), m)
	if m != nil {
		switch m.MType {
		case metrics.MTypeGauge:
			v := 123.45
			m.Value = &v
		case metrics.MTypeCounter:
			d := int64(12345)
			m.Delta = &d
		}
	}
	return retArgs.Error(0)
}

func TestReadByJSONHandler(t *testing.T) {
	makeURL := func(serverURL string) string {
		return serverURL + "/value/"
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

		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &metrics.Metrics{},
		).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{"id": "0", "type": "gauge"}`,
			)

			req, err := http.NewRequestWithContext(
				context.Background(), method, makeURL(s.URL), reqBody,
			)
			require.NoError(t, err)
			req.Header.Set(httpapi.ContentType, httpapi.JSON)

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Empty(t, data)

			mockService.AssertNumberOfCalls(t, "Read", 0)
		}
	})

	t.Run("Not allowed Content-Type", func(t *testing.T) {
		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &metrics.Metrics{},
		).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{"id": "0", "type": "gauge"}`,
		)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(httpapi.ContentType, httpapi.TEXT)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)
		assert.Empty(t, data)

		mockService.AssertNumberOfCalls(t, "Read", 0)
	})

	t.Run("Bad JSON", func(t *testing.T) {
		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &metrics.Metrics{},
		).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{"id": "0", "type": "gauge,}`,
		)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(httpapi.ContentType, httpapi.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		mockService.AssertNumberOfCalls(t, "Read", 0)
	})

	t.Run("Invalid metrics payload", func(t *testing.T) {
		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &metrics.Metrics{},
		).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{ID: "", MType: ""}`,
		)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(httpapi.ContentType, httpapi.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		res.Body.Close()
		mockService.AssertNumberOfCalls(t, "Read", 0)
	})

	t.Run("Not exists error on Read", func(t *testing.T) {
		var schemeReq metrics.Metrics
		schemeReq.ID = "0"
		schemeReq.MType = metrics.MTypeGauge

		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &schemeReq,
		).Return(server.ErrNotExists)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(httpapi.ContentType, httpapi.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, server.ErrNotExists.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 1)
	})

	t.Run("Internal error on Read", func(t *testing.T) {
		var schemeReq metrics.Metrics
		schemeReq.ID = "0"
		schemeReq.MType = metrics.MTypeGauge
		readErr := errors.New("test error")

		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &schemeReq,
		).Return(readErr)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(httpapi.ContentType, httpapi.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, server.ErrInternal.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		var schemeReq metrics.Metrics
		schemeReq.ID = "0"
		schemeReq.MType = metrics.MTypeGauge

		mockService := new(MockValueService)
		mockService.On("Read", context.Background(), &schemeReq).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		var bufReq bytes.Buffer
		err := json.NewEncoder(&bufReq).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodPost, makeURL(s.URL), &bufReq,
		)
		require.NoError(t, err)
		req.Header.Set(httpapi.ContentType, httpapi.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, httpapi.JSON, res.Header.Get(httpapi.ContentType))

		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(
			t,
			`{"id": "0", "type": "gauge", "value": 123.45}`,
			string(data),
		)
		res.Body.Close()

		mockService.AssertNumberOfCalls(t, "Read", 1)
	})

	t.Run("Encoding", func(t *testing.T) {
		t.Run("Allow only gzip", func(t *testing.T) {
			mockService := new(MockValueService)
			mockService.On(
				"Read", context.Background(), &metrics.Metrics{},
			).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)
			httpapi.SetValueHandler(mux, mockService)

			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{id: "0", type: "gauge"}`,
			)

			req, err := http.NewRequestWithContext(
				context.Background(), http.MethodPost, makeURL(s.URL), reqBody,
			)
			require.NoError(t, err)
			req.Header.Set(httpapi.ContentType, httpapi.JSON)
			req.Header.Set(httpapi.ContentEncoding, "br")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Empty(t, data)

			mockService.AssertNumberOfCalls(t, "Read", 0)
		})

		t.Run("Send gzip, accept non-compressed", func(t *testing.T) {
			var schemeReq metrics.Metrics
			schemeReq.ID = "0"
			schemeReq.MType = metrics.MTypeGauge

			mockService := new(MockValueService)
			mockService.On(
				"Read", context.Background(), &schemeReq,
			).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)
			httpapi.SetValueHandler(mux, mockService)

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
			req.Header.Set(httpapi.ContentType, httpapi.JSON)
			req.Header.Set(httpapi.ContentEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			assert.Empty(t, res.Header.Get(httpapi.ContentEncoding))

			data, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.JSONEq(
				t,
				`{"id": "0", "type": "gauge", "value": 123.45}`,
				string(data),
			)
			res.Body.Close()

			mockService.AssertNumberOfCalls(t, "Read", 1)
		})

		t.Run("Send non-compressed accept gzip", func(t *testing.T) {
			var schemeReq metrics.Metrics
			schemeReq.ID = "0"
			schemeReq.MType = metrics.MTypeGauge

			mockService := new(MockValueService)
			mockService.On(
				"Read", context.Background(), &schemeReq,
			).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)
			httpapi.SetValueHandler(mux, mockService)

			s := httptest.NewServer(mux)
			defer s.Close()

			var bufReq bytes.Buffer
			err := json.NewEncoder(&bufReq).Encode(&schemeReq)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				context.Background(), http.MethodPost, makeURL(s.URL), &bufReq,
			)
			require.NoError(t, err)
			req.Header.Set(httpapi.ContentType, httpapi.JSON)
			req.Header.Set(httpapi.AcceptEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			gzipReader, err := gzip.NewReader(res.Body)
			require.NoError(t, err)
			data, err := io.ReadAll(gzipReader)
			gzipReader.Close()
			res.Body.Close()
			require.NoError(t, err)
			expect := `{"id":"0","type":"gauge","value":123.45}`
			assert.JSONEq(t, expect, string(data))

			mockService.AssertNumberOfCalls(t, "Read", 1)
		})
	})
}

func TestReadByURLParamsHandler(t *testing.T) {
	makeURL := func(serverURL, mType, mName string) string {
		testURL, _ := url.JoinPath(serverURL+"/value", mType, mName)
		return testURL
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
		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &metrics.Metrics{},
		).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			req, err := http.NewRequestWithContext(
				context.Background(),
				method,
				makeURL(s.URL, "gauge", "Alloc"),
				http.NoBody,
			)
			require.NoError(t, err)

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Empty(t, data)

			mockService.AssertNumberOfCalls(t, "Read", 0)
		}
	})

	t.Run("Invalid metrics payload", func(t *testing.T) {
		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &metrics.Metrics{},
		).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			makeURL(s.URL, "invalidType", "0"),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		res.Body.Close()
		mockService.AssertNumberOfCalls(t, "Read", 0)
	})

	t.Run("Not exists error on Read", func(t *testing.T) {
		id := "Alloc"
		mType := metrics.MTypeGauge

		var schemeReq metrics.Metrics
		schemeReq.ID = id
		schemeReq.MType = mType

		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &schemeReq,
		).Return(server.ErrNotExists)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			makeURL(s.URL, mType, id),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, server.ErrNotExists.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 1)
	})

	t.Run("Internal error on Read", func(t *testing.T) {
		id := "Alloc"
		mType := metrics.MTypeGauge
		readErr := errors.New("test error")

		var schemeReq metrics.Metrics
		schemeReq.ID = id
		schemeReq.MType = mType

		mockService := new(MockValueService)
		mockService.On(
			"Read", context.Background(), &schemeReq,
		).Return(readErr)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			makeURL(s.URL, mType, id),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, server.ErrInternal.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeGauge

		var schemeReq metrics.Metrics
		schemeReq.ID = id
		schemeReq.MType = mType

		mockService := new(MockValueService)
		mockService.On("Read", context.Background(), &schemeReq).Return(nil)

		mux := chi.NewRouter()
		httpapi.SetValueHandler(mux, mockService)

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			makeURL(s.URL, mType, id),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		expect := `123.45`
		assert.Equal(t, expect, string(data))

		mockService.AssertNumberOfCalls(t, "Read", 1)
	})

}
