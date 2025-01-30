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
	io.ReadCloser
	gzip *gzip.Reader
}

func newGzipReader(requestBody io.ReadCloser) (*gzipReader, error) {
	r, err := gzip.NewReader(requestBody)
	if err != nil {
		return nil, err
	}
	return &gzipReader{ReadCloser: requestBody, gzip: r}, nil
}

func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.gzip.Read(p)
}

func (r *gzipReader) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return err
	}

	return r.gzip.Close()
}

type gzipWriter struct {
	http.ResponseWriter
	gzip         *gzip.Writer
	contentTypes map[string]struct{}
	compressable bool
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	gzip, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	contentTypes := map[string]struct{}{
		"application/json": {},
		"text/html":        {},
	}
	return &gzipWriter{ResponseWriter: w, gzip: gzip, contentTypes: contentTypes}
}

func (w *gzipWriter) isCompressable() bool {
	contentType := strings.Split(w.Header().Get(ContentType), ";")[0]
	_, ok := w.contentTypes[contentType]
	return ok
}

func (w *gzipWriter) Write(p []byte) (int, error) {
	if !w.compressable {
		return w.ResponseWriter.Write(p)
	}
	return w.gzip.Write(p)
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 && w.isCompressable() {
		w.compressable = true
		w.Header().Set(ContentEncoding, gzipFormat)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipWriter) Close() error {
	if w.compressable {
		return w.gzip.Close()
	}
	return nil
}

func receiveGzip(reqHeader *http.Header) bool {
	for _, reqEncoding := range reqHeader.Values(ContentEncoding) {
		reqEncoding = strings.ToLower(strings.TrimSpace(reqEncoding))
		if reqEncoding == gzipFormat {
			return true
		}
	}

	return false
}

func acceptGzip(reqHeader *http.Header) bool {
	acceptEncodings := reqHeader.Values(AcceptEncoding)
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
