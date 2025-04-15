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
	ctx context.Context, scheme *metrics.Metrics,
) error {
	retArgs := service.Called(ctx, scheme)
	return retArgs.Error(0)
}

type MockUpdateValidator struct {
	mock.Mock
}

func (validator *MockUpdateValidator) VerifyScheme(
	verifier di.Verifier,
) error {
	retArgs := validator.Called(verifier)
	return retArgs.Error(0)
}

func (validator *MockUpdateValidator) VerifyParams(
	id, mType, value string,
) (metrics.Metrics, error) {
	retArgs := validator.Called(id, mType, value)
	return retArgs.Get(0).(metrics.Metrics), retArgs.Error(1)
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
		mockService.On("Update", nil, nil).Return(nil)
		mockValidator := new(MockUpdateValidator)
		mockValidator.On("VerifyScheme", nil).Return(nil)
		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{ID: "0", MType: "gauge", Value: 123.450}`,
			)

			req, err := http.NewRequestWithContext(
				context.TODO(), method, makeURL(s.URL), reqBody,
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
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
		}
	})

	t.Run("Not allowed Content-Type", func(t *testing.T) {
		mockService := new(MockUpdateService)
		mockService.On("Update", nil, nil).Return(nil)
		mockValidator := new(MockUpdateValidator)
		mockValidator.On("VerifyScheme", nil).Return(nil)
		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{ID: "0", MType: "gauge", Value: 123.450}`,
		)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, "text/plain")

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)
		assert.Len(t, data, 0)

		mockService.AssertNumberOfCalls(t, "Update", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
	})

	t.Run("Bad JSON", func(t *testing.T) {
		mockService := new(MockUpdateService)
		mockService.On("Update", nil, nil).Return(nil)
		mockValidator := new(MockUpdateValidator)
		mockValidator.On("VerifyScheme", nil).Return(nil)
		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)

		s := httptest.NewServer(mux)
		defer s.Close()

		reqBody := strings.NewReader(
			`{ID: "0", MType: "gauge, Value: 123.450}`,
		)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), reqBody,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		mockService.AssertNumberOfCalls(t, "Update", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
	})

	t.Run("Error on verifyScheme", func(t *testing.T) {
		var schemeReq metrics.MetricsUpdate
		schemeReq.ID = ""
		schemeReq.MType = metrics.MTypeGauge
		expectedErr := errors.New("test error")
		mockService := new(MockUpdateService)
		mockService.On("Update", nil, nil).Return(nil)
		mockValidator := new(MockUpdateValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(expectedErr)
		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, expectedErr.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Update", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Error on Update", func(t *testing.T) {
		value := 123.45
		var schemeReq metrics.MetricsUpdate
		schemeReq.ID = "0"
		schemeReq.MType = metrics.MTypeGauge
		schemeReq.Value = &value
		expectedErr := errors.New("test error")
		mockService := new(MockUpdateService)
		mockService.On("Update", nil, &schemeReq).Return(expectedErr)
		mockValidator := new(MockUpdateValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)

		s := httptest.NewServer(mux)
		defer s.Close()

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), &buf,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, expectedErr.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Update", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeGauge
		value := 123.45

		var schemeReq metrics.MetricsUpdate
		schemeReq.ID = id
		schemeReq.MType = mType
		schemeReq.Value = &value

		var schemeRes metrics.Metrics
		schemeRes.ID = id
		schemeRes.MType = mType
		schemeRes.Value = &value

		mockService := new(MockUpdateService)
		mockService.On("Update", nil, &schemeReq).Return(nil)
		mockValidator := new(MockUpdateValidator)
		mockValidator.On("VerifyScheme", &schemeReq).Return(nil)
		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)

		s := httptest.NewServer(mux)
		defer s.Close()

		var bufReq bytes.Buffer
		err := json.NewEncoder(&bufReq).Encode(schemeReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			context.TODO(), http.MethodPost, makeURL(s.URL), &bufReq,
		)
		require.NoError(t, err)
		req.Header.Set(header.ContentType, mime.JSON)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		var gotScheme metrics.Metrics
		err = json.NewDecoder(res.Body).Decode(&gotScheme)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, schemeRes, gotScheme)

		mockService.AssertNumberOfCalls(t, "Update", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
	})

	t.Run("Encoding", func(t *testing.T) {
		t.Run("Allow only gzip", func(t *testing.T) {
			mockService := new(MockUpdateService)
			mockService.On("Update", nil, nil).Return(nil)

			mockValidator := new(MockUpdateValidator)
			mockValidator.On("VerifyScheme", nil).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)

			SetUpdateHandler(mux, mockService, mockValidator)

			s := httptest.NewServer(mux)
			defer s.Close()

			reqBody := strings.NewReader(
				`{ID: "0", MType: "gauge", Value: 123.450}`,
			)

			req, err := http.NewRequestWithContext(
				context.TODO(), http.MethodPost, makeURL(s.URL), reqBody,
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
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 0)
		})

		t.Run("Send gzip, accept non-compressed", func(t *testing.T) {
			id := "0"
			mType := metrics.MTypeGauge
			value := 123.45

			var schemeReq metrics.MetricsUpdate
			schemeReq.ID = id
			schemeReq.MType = mType
			schemeReq.Value = &value

			var schemeRes metrics.Metrics
			schemeRes.ID = id
			schemeRes.MType = mType
			schemeRes.Value = &value

			mockService := new(MockUpdateService)
			mockService.On("Update", nil, &schemeReq).Return(nil)

			mockValidator := new(MockUpdateValidator)
			mockValidator.On("VerifyScheme", &schemeReq).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)

			SetUpdateHandler(mux, mockService, mockValidator)

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
			req.Header.Set(header.ContentType, mime.JSON)
			req.Header.Set(header.ContentEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			assert.Zero(t, res.Header.Get(header.ContentEncoding))

			var schemeGot metrics.Metrics
			err = json.NewDecoder(res.Body).Decode(&schemeGot)
			res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, schemeRes, schemeGot)
			mockService.AssertNumberOfCalls(t, "Update", 1)
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
		})

		t.Run("Send non-compressed, accept gzip", func(t *testing.T) {
			id := "0"
			mType := metrics.MTypeGauge
			value := 123.45

			var schemeReq metrics.MetricsUpdate
			schemeReq.ID = id
			schemeReq.MType = mType
			schemeReq.Value = &value

			var schemeRes metrics.Metrics
			schemeRes.ID = id
			schemeRes.MType = mType
			schemeRes.Value = &value

			mockService := new(MockUpdateService)
			mockService.On("Update", nil, &schemeReq).Return(nil)

			mockValidator := new(MockUpdateValidator)
			mockValidator.On("VerifyScheme", &schemeReq).Return(nil)

			mux := chi.NewRouter()
			mux.Use(middleware.AllowContentEncoding("gzip"))
			mux.Use(middleware.Gzip)

			SetUpdateHandler(mux, mockService, mockValidator)

			s := httptest.NewServer(mux)
			defer s.Close()

			var bufReq bytes.Buffer
			err := json.NewEncoder(&bufReq).Encode(&schemeReq)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				context.TODO(), http.MethodPost, makeURL(s.URL), &bufReq,
			)
			require.NoError(t, err)
			req.Header.Set(header.ContentType, mime.JSON)
			req.Header.Set(header.AcceptEncoding, "gzip")

			res, err := s.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			var schemeGot metrics.Metrics
			gzipReader, err := gzip.NewReader(res.Body)
			require.NoError(t, err)
			err = json.NewDecoder(gzipReader).Decode(&schemeGot)
			res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, schemeRes, schemeGot)
			mockService.AssertNumberOfCalls(t, "Update", 1)
			mockValidator.AssertNumberOfCalls(t, "VerifyScheme", 1)
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
		mockService.On("Update", nil, nil).Return(nil)

		mockValidator := new(MockUpdateValidator)
		mockValidator.On(
			"VerifyParams", "0", "gauge", "123.450",
		).Return(nil, nil)

		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)

		for _, method := range methods {
			s := httptest.NewServer(mux)
			defer s.Close()

			req, err := http.NewRequestWithContext(
				context.TODO(),
				method,
				makeURL(s.URL, "gauge", "Alloc", "123.450"),
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

	t.Run("Error on verify params", func(t *testing.T) {
		id := "0"
		mType := "bgauge"
		value := "123.450"
		expectedErr := errors.New("test error")

		mockService := new(MockUpdateService)
		mockService.On("Update", nil, nil).Return(nil)

		mockValidator := new(MockUpdateValidator)
		mockValidator.On(
			"VerifyParams", id, mType, value,
		).Return(metrics.Metrics{}, expectedErr)

		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)
		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.TODO(),
			http.MethodPost,
			makeURL(s.URL, mType, id, value),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, expectedErr.Error(), strings.TrimSpace(string(data)))

		mockService.AssertNumberOfCalls(t, "Update", 0)
		mockValidator.AssertNumberOfCalls(t, "VerifyParams", 1)
	})

	t.Run("Error on update", func(t *testing.T) {
		id := "0"
		mType := "gauge"
		value := "123.450"
		valueFloat := 123.450
		errOnUpdate := errors.New("test error")

		var schemeUpdate metrics.Metrics
		schemeUpdate.ID = id
		schemeUpdate.MType = mType
		schemeUpdate.Value = &valueFloat

		mockService := new(MockUpdateService)
		mockService.On("Update", nil, &schemeUpdate).Return(errOnUpdate)

		mockValidator := new(MockUpdateValidator)
		mockValidator.On(
			"VerifyParams", id, mType, value,
		).Return(schemeUpdate, nil)

		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)
		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.TODO(),
			http.MethodPost,
			makeURL(s.URL, mType, id, value),
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
		mockValidator.AssertNumberOfCalls(t, "VerifyParams", 1)
	})

	t.Run("Regular response", func(t *testing.T) {
		id := "0"
		mType := "gauge"
		value := "123.45"
		schemeValue := 123.45

		var schemeReq metrics.Metrics
		schemeReq.ID = id
		schemeReq.MType = mType
		schemeReq.Value = &schemeValue

		mockService := new(MockUpdateService)
		mockService.On("Update", nil, &schemeReq).Return(nil)

		mockValidator := new(MockUpdateValidator)
		mockValidator.On(
			"VerifyParams", id, mType, value,
		).Return(schemeReq, nil)

		mux := chi.NewRouter()

		SetUpdateHandler(mux, mockService, mockValidator)
		s := httptest.NewServer(mux)
		defer s.Close()

		req, err := http.NewRequestWithContext(
			context.TODO(),
			http.MethodPost,
			makeURL(s.URL, mType, id, value),
			http.NoBody,
		)
		require.NoError(t, err)

		res, err := s.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, value, string(data))

		mockService.AssertNumberOfCalls(t, "Update", 1)
		mockValidator.AssertNumberOfCalls(t, "VerifyParams", 1)
	})
}
