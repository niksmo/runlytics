package api

// type MockValueService struct {
// 	err bool
// }

// func (service *MockValueService) Read(mData *metrics.Metrics) error {
// 	if service.err {
// 		return errors.New("test error")

// 	}

// 	value := 123.4
// 	mData.Value = &value

// 	delta := int64(1234)
// 	mData.Delta = &delta

// 	return nil
// }

// func TestReadByJSONHandler(t *testing.T) {
// 	newMetrics := func(id string, mType string, delta int64, value float64) *metrics.Metrics {
// 		return &metrics.Metrics{
// 			ID:    id,
// 			MType: mType,
// 			Delta: &delta,
// 			Value: &value,
// 		}
// 	}

// 	testRequest := func(
// 		t *testing.T,
// 		mux *chi.Mux,
// 		method string,
// 		path string,
// 		contentType string,
// 		body io.Reader,
// 	) (*http.Response, []byte) {
// 		ts := httptest.NewServer(mux)
// 		defer ts.Close()

// 		req, err := http.NewRequest(method, ts.URL+path, body)
// 		require.NoError(t, err)

// 		req.Header.Set("Content-Type", contentType)
// 		res, err := ts.Client().Do(req)
// 		require.NoError(t, err)

// 		resData, err := io.ReadAll(res.Body)
// 		require.NoError(t, err)

// 		return res, resData
// 	}

// 	type want struct {
// 		statusCode int
// 		resData    *metrics.Metrics
// 	}

// 	type test struct {
// 		name        string
// 		method      string
// 		path        string
// 		want        want
// 		contentType string
// 		reqData     []byte
// 		service     *MockValueService
// 	}

// 	tests := []test{
// 		// not allowed methods
// 		{
// 			name:   "GET not allowed",
// 			method: http.MethodGet,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 				resData:    nil,
// 			},
// 		},
// 		{
// 			name:   "PUT not allowed",
// 			method: http.MethodPut,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 				resData:    nil,
// 			},
// 		},
// 		{
// 			name:   "PATCH not allowed",
// 			method: http.MethodPatch,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 				resData:    nil,
// 			},
// 		},
// 		{
// 			name:   "DELETE not allowed",
// 			method: http.MethodDelete,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 				resData:    nil,
// 			},
// 		},
// 		{
// 			name:   "HEAD not allowed",
// 			method: http.MethodHead,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 				resData:    nil,
// 			},
// 		},
// 		{
// 			name:   "OPTIONS not allowed",
// 			method: http.MethodOptions,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 				resData:    nil,
// 			},
// 		},

// 		//POST
// 		{
// 			name:   "Should read metrics",
// 			method: http.MethodPost,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusOK,
// 				resData: newMetrics(
// 					"test",
// 					"gauge",
// 					int64(1234),
// 					123.4,
// 				),
// 			},
// 			contentType: "application/json",
// 			reqData:     []byte(`{"id": "test", "type": "gauge"}`),
// 			service:     &MockValueService{err: false},
// 		},
// 		{
// 			name:   "Wrong Content-Type",
// 			method: http.MethodPost,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusUnsupportedMediaType,
// 				resData:    nil,
// 			},
// 			contentType: "text/plain",
// 			reqData:     []byte(`{"id": "test", "type": "gauge"}`),
// 			service:     &MockValueService{err: false},
// 		},
// 		{
// 			name:   "Wrong metrics scheme or bad JSON",
// 			method: http.MethodPost,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusBadRequest,
// 				resData:    nil,
// 			},
// 			contentType: "application/json",
// 			reqData:     []byte(`{"id": "test", "type": "gauge}`),
// 			service:     &MockValueService{err: false},
// 		},
// 		{
// 			name:   "Service should return error",
// 			method: http.MethodPost,
// 			path:   "/value/",
// 			want: want{
// 				statusCode: http.StatusNotFound,
// 				resData:    nil,
// 			},
// 			contentType: "application/json",
// 			reqData:     []byte(`{"id": "test", "type": "gauge"}`),
// 			service:     &MockValueService{err: true},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			mux := chi.NewRouter()
// 			SetReadHandler(mux, test.service)
// 			res, resBody := testRequest(t,
// 				mux,
// 				test.method,
// 				test.path,
// 				test.contentType,
// 				bytes.NewReader(test.reqData),
// 			)
// 			defer res.Body.Close()

// 			assert.Equal(t, test.want.statusCode, res.StatusCode)

// 			if test.want.resData != nil {
// 				var resData metrics.Metrics
// 				require.NoError(t, json.Unmarshal(resBody, &resData))
// 				assert.Equal(t, *test.want.resData, resData)
// 			}
// 		})
// 	}
// }

// func TestReadByURLParamsHandler(t *testing.T) {

// 	testRequest := func(
// 		t *testing.T,
// 		mux *chi.Mux,
// 		method string,
// 		path string,
// 		body io.Reader,
// 	) (*http.Response, []byte) {
// 		ts := httptest.NewServer(mux)
// 		defer ts.Close()

// 		req, err := http.NewRequest(method, ts.URL+path, body)
// 		require.NoError(t, err)

// 		req.Header.Set("Accept-Encoding", "gzip")

// 		res, err := ts.Client().Do(req)
// 		require.NoError(t, err)

// 		resData, err := io.ReadAll(res.Body)
// 		require.NoError(t, err)

// 		return res, resData
// 	}

// 	type want struct {
// 		statusCode int
// 		resData    string
// 	}

// 	type test struct {
// 		name    string
// 		method  string
// 		path    string
// 		want    want
// 		service *MockValueService
// 	}

// 	tests := []test{
// 		// not allowed methods
// 		{
// 			name:   "POST not allowed",
// 			method: http.MethodPost,
// 			path:   "/value/gauge/testName",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 			},
// 		},
// 		{
// 			name:   "PUT not allowed",
// 			method: http.MethodPut,
// 			path:   "/value/gauge/testName",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 			},
// 		},
// 		{
// 			name:   "PATCH not allowed",
// 			method: http.MethodPatch,
// 			path:   "/value/gauge/testName",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 			},
// 		},
// 		{
// 			name:   "DELETE not allowed",
// 			method: http.MethodDelete,
// 			path:   "/value/gauge/testName",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 			},
// 		},
// 		{
// 			name:   "HEAD not allowed",
// 			method: http.MethodHead,
// 			path:   "/value/gauge/testName",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 			},
// 		},
// 		{
// 			name:   "OPTIONS not allowed",
// 			method: http.MethodOptions,
// 			path:   "/value/gauge/testName",
// 			want: want{
// 				statusCode: http.StatusMethodNotAllowed,
// 			},
// 		},

// 		// GET
// 		{
// 			name:   "Should read metrics",
// 			method: http.MethodGet,
// 			path:   "/value/gauge/testName",
// 			want: want{
// 				statusCode: http.StatusOK,
// 				resData:    "123.4",
// 			},
// 			service: &MockValueService{err: false},
// 		},
// 		{
// 			name:   "Service should return error",
// 			method: http.MethodGet,
// 			want: want{
// 				statusCode: http.StatusNotFound,
// 			},
// 			service: &MockValueService{err: true},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			mux := chi.NewRouter()
// 			mux.Use(middleware.AllowContentEncoding("gzip"))
// 			mux.Use(middleware.Gzip)
// 			SetReadHandler(mux, test.service)
// 			res, resBody := testRequest(t,
// 				mux,
// 				test.method,
// 				test.path,
// 				http.NoBody,
// 			)
// 			defer res.Body.Close()

// 			assert.Equal(t, test.want.statusCode, res.StatusCode)

// 			if test.want.resData != "" {
// 				require.Empty(t, res.Header.Get("Content-Encoding"))
// 				assert.Equal(t, test.want.resData, string(resBody))
// 			}
// 		})
// 	}
// }
