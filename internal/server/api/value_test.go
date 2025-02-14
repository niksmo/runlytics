package api

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
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockValueService struct {
	mock.Mock
}

func (service *MockValueService) Read(
	ctx context.Context, scheme *metrics.MetricsRead,
) (di.Metrics, error) {
	retArgs := service.Called(nil, scheme)

	if _, ok := retArgs.Get(0).(di.Metrics); !ok {
		return nil, retArgs.Error(1)
	}
	return retArgs.Get(0).(di.Metrics), retArgs.Error(1)
}

type MockValueValidator struct {
	mock.Mock
}

func (validator *MockValueValidator) VerifyScheme(
	verifier di.Verifier,
) error {
	retArgs := validator.Called(verifier)
	return retArgs.Error(0)
}

type MockValueConfig struct{}

func (config *MockValueConfig) Key() string {
	return ""
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
		mockService.On("Read", nil, nil).Return(nil, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", nil).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{ID: "0", MType: "gauge"}`,
			)

			req, err := http.NewRequestWithContext(
				context.TODO(), method, makeURL(s.URL), reqBody,
			)
			require.NoError(t, err)
			req.Header.Set(ContentType, JSONMediaType)

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Len(t, data, 0)

			mockService.AssertNumberOfCalls(t, "Read", 0)
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
		}
	})

	t.Run("Not allowed Content-Type", func(t *testing.T) {
		mockService := new(MockValueService)
		mockService.On("Read", nil, nil).Return(nil, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", nil).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{ID: "0", MType: "gauge"}`,
		)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(ContentType, "text/plain")

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)
		assert.Len(t, data, 0)

		mockService.AssertNumberOfCalls(t, "Read", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
	})

	t.Run("Bad JSON", func(t *testing.T) {
		mockService := new(MockValueService)
		mockService.On("Read", nil, nil).Return(nil, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", nil).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{ID: "0", MType: "gauge}`,
		)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(ContentType, JSONMediaType)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		mockService.AssertNumberOfCalls(t, "Read", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
	})

	t.Run("Error on verifyScheme", func(t *testing.T) {
		schemeReq := metrics.MetricsRead{
			ID: "", MType: metrics.MTypeGauge,
		}
		expectedErr := errors.New("test error")
		mockService := new(MockValueService)
		mockService.On("Read", nil, nil).Return(nil, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(expectedErr)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(ContentType, JSONMediaType)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, expectedErr.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Not exists error on Read", func(t *testing.T) {
		schemeReq := metrics.MetricsRead{
			ID: "0", MType: metrics.MTypeGauge,
		}
		expectedErr := server.ErrNotExists
		mockService := new(MockValueService)
		mockService.On("Read", nil, &schemeReq).Return(nil, expectedErr)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(ContentType, JSONMediaType)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, expectedErr.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Internal error on Read", func(t *testing.T) {
		schemeReq := metrics.MetricsRead{
			ID: "0", MType: metrics.MTypeGauge,
		}
		mockService := new(MockValueService)
		mockService.On("Read", nil, &schemeReq).Return(nil, errors.New("test error"))
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(ContentType, JSONMediaType)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, server.ErrInternal.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeGauge
		value := 123.45
		schemeReq := metrics.MetricsRead{
			ID: id, MType: mType,
		}
		schemeRes := metrics.MetricsGauge{
			ID: id, MType: mType, Value: value,
		}
		mockService := new(MockValueService)
		mockService.On("Read", nil, &schemeReq).Return(schemeRes, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		var bufReq bytes.Buffer
		err := json.NewEncoder(&bufReq).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), &bufReq,
		)
		require.NoError(t, err)
		req.Header.Set(ContentType, JSONMediaType)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		var gotScheme metrics.MetricsGauge
		err = json.NewDecoder(res.Body).Decode(&gotScheme)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, schemeRes, gotScheme)

		mockService.AssertNumberOfCalls(t, "Read", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Encoding", func(t *testing.T) {
		t.Run("Allow only gzip", func(t *testing.T) {
			mockService := new(MockValueService)
			mockService.On("Read", nil, nil).Return(nil, nil)
			mockValidator := new(MockValueValidator)
			mockValidator.On("VerifyScheme", nil).Return(nil)
			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)

			SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{ID: "0", MType: "gauge"}`,
			)

			req, err := http.NewRequestWithContext(
				context.TODO(), http.MethodPost, makeURL(s.URL), reqBody,
			)
			require.NoError(t, err)
			req.Header.Set(ContentType, JSONMediaType)
			req.Header.Set(ContentEncoding, "br")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Len(t, data, 0)

			mockService.AssertNumberOfCalls(t, "Read", 0)
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
		})

		t.Run("Send gzip, accept non-compressed", func(t *testing.T) {
			id := "0"
			mType := metrics.MTypeGauge
			value := 123.45
			schemeReq := metrics.MetricsRead{
				ID: id, MType: mType,
			}
			schemeRes := metrics.MetricsGauge{
				ID: id, MType: mType, Value: value,
			}

			mockValidator := new(MockValueValidator)
			mockValidator.On("VerifyScheme", &schemeReq).Return(nil)

			mockService := new(MockValueService)
			mockService.On("Read", nil, &schemeReq).Return(schemeRes, nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)

			SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

			s := httptest.NewServer(mux)
			defer s.Close()

			var buf bytes.Buffer
			gzipWriter := gzip.NewWriter(&buf)
			err := json.NewEncoder(gzipWriter).Encode(&schemeReq)
			require.NoError(t, err)
			err = gzipWriter.Close()
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				context.TODO(), http.MethodPost, makeURL(s.URL), &buf,
			)
			require.NoError(t, err)
			req.Header.Set(ContentType, JSONMediaType)
			req.Header.Set(ContentEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			assert.Zero(t, res.Header.Get(ContentEncoding))

			var schemeGot metrics.MetricsGauge
			err = json.NewDecoder(res.Body).Decode(&schemeGot)
			res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, schemeRes, schemeGot)
			mockService.AssertNumberOfCalls(t, "Read", 1)
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
		})

		t.Run("Send non-compressed, accept gzip", func(t *testing.T) {
			id := "0"
			mType := metrics.MTypeGauge
			value := 123.45
			schemeReq := metrics.MetricsRead{
				ID: id, MType: mType,
			}
			schemeRes := metrics.MetricsGauge{
				ID: id, MType: mType, Value: value,
			}

			mockValidator := new(MockValueValidator)
			mockValidator.On("VerifyScheme", &schemeReq).Return(nil)

			mockService := new(MockValueService)
			mockService.On("Read", nil, &schemeReq).Return(schemeRes, nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)

			SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

			s := httptest.NewServer(mux)
			defer s.Close()

			var bufReq bytes.Buffer
			err := json.NewEncoder(&bufReq).Encode(&schemeReq)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				context.TODO(), http.MethodPost, makeURL(s.URL), &bufReq,
			)
			require.NoError(t, err)
			req.Header.Set(ContentType, JSONMediaType)
			req.Header.Set(AcceptEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			var schemeGot metrics.MetricsGauge
			gzipReader, err := gzip.NewReader(res.Body)
			require.NoError(t, err)
			err = json.NewDecoder(gzipReader).Decode(&schemeGot)
			res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, schemeRes, schemeGot)
			mockService.AssertNumberOfCalls(t, "Read", 1)
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
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
		mockService.On("Update", nil, nil).Return(nil, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", nil).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			req, err := http.NewRequestWithContext(
				context.TODO(),
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
			assert.Len(t, data, 0)

			mockService.AssertNumberOfCalls(t, "Update", 0)
			mockValidator.AssertNumberOfCalls(t, "VerifyParams", 0)
		}
	})

	t.Run("Error on verify scheme", func(t *testing.T) {
		id := "Alloc"
		mType := metrics.MTypeGauge
		schemeReq := metrics.MetricsRead{
			ID: id, MType: mType,
		}
		expectedErr := errors.New("test error")
		mockService := new(MockValueService)
		mockService.On("Read", nil, nil).Return(nil, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(expectedErr)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.TODO(),
			http.MethodGet,
			makeURL(s.URL, mType, id),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, expectedErr.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Not exists error on Read", func(t *testing.T) {
		id := "Alloc"
		mType := metrics.MTypeGauge
		schemeReq := metrics.MetricsRead{
			ID: id, MType: mType,
		}
		expectedErr := server.ErrNotExists
		mockService := new(MockValueService)
		mockService.On("Read", nil, &schemeReq).Return(nil, expectedErr)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.TODO(),
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
		assert.Equal(t, expectedErr.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Read", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Internal error on Read", func(t *testing.T) {
		id := "Alloc"
		mType := metrics.MTypeGauge
		schemeReq := metrics.MetricsRead{
			ID: id, MType: mType,
		}
		mockService := new(MockValueService)
		mockService.On("Read", nil, &schemeReq).Return(nil, errors.New("test error"))
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.TODO(),
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
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeGauge
		value := 123.45
		schemeReq := metrics.MetricsRead{
			ID: id, MType: mType,
		}
		schemeRes := metrics.MetricsGauge{
			ID: id, MType: mType, Value: value,
		}
		mockService := new(MockValueService)
		mockService.On("Read", nil, &schemeReq).Return(schemeRes, nil)
		mockValidator := new(MockValueValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetValueHandler(mux, mockService, mockValidator, new(MockValueConfig))

		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.TODO(),
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
		assert.Equal(t, "123.45", string(data))

		mockService.AssertNumberOfCalls(t, "Read", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

}
