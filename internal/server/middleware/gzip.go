package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

const gzipFormat = "gzip"

func Gzip(next http.Handler) http.Handler {
	gzipFunc := func(w http.ResponseWriter, r *http.Request) {

		if receiveGzip(&r.Header) && r.ContentLength > 0 {
			gzipR, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = gzipR
			logger.Log.Debug("Receive compressed data", zap.String("format", gzipFormat))
		}

		if acceptGzip(&r.Header) {
			gzipW := newGzipWriter(w)
			defer gzipW.Close()
			w = gzipW
			logger.Log.Debug("Prepare compressed response", zap.String("format", gzipFormat))
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(gzipFunc)
}

type gzipReader struct {
	wrap io.ReadCloser
	gzip *gzip.Reader
}

func newGzipReader(requestBody io.ReadCloser) (*gzipReader, error) {
	r, err := gzip.NewReader(requestBody)
	if err != nil {
		return nil, err
	}
	return &gzipReader{wrap: requestBody, gzip: r}, nil
}

func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.gzip.Read(p)
}

func (r *gzipReader) Close() error {
	if err := r.wrap.Close(); err != nil {
		return err
	}

	return r.gzip.Close()
}

type gzipWriter struct {
	wrap http.ResponseWriter
	gzip *gzip.Writer
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	gzip, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	return &gzipWriter{wrap: w, gzip: gzip}
}

func (w *gzipWriter) Header() http.Header {
	return w.wrap.Header()
}

func (w *gzipWriter) Write(p []byte) (int, error) {
	return w.gzip.Write(p)
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		w.wrap.Header().Set("Content-Encoding", gzipFormat)
	}
	w.wrap.WriteHeader(statusCode)
}

func (w *gzipWriter) Close() error {
	return w.gzip.Close()
}

func receiveGzip(reqHeader *http.Header) bool {
	for _, reqEncoding := range reqHeader.Values("Content-Encoding") {
		reqEncoding = strings.ToLower(strings.TrimSpace(reqEncoding))
		if reqEncoding == gzipFormat {
			return true
		}
	}

	return false
}

func acceptGzip(reqHeader *http.Header) bool {
	acceptEncodings := reqHeader.Values("Accept-Encoding")
	for _, acceptEncoding := range acceptEncodings {
		if strings.HasPrefix(acceptEncoding, "*") {
			return true
		}

		if strings.HasPrefix(strings.ToLower(acceptEncoding), gzipFormat) {
			return true
		}
	}
	return false
}
