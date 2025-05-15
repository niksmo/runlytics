package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/cipher"
	"go.uber.org/zap"
)

func Decrypt(PEMData []byte) func(http.Handler) http.Handler {
	decrypter, err := cipher.NewDecrypter(PEMData)
	if err != nil {
		logger.Log.Fatal("failed to init decrypter", zap.Error(err))
	}

	middlewareFunc := func(next http.Handler) http.Handler {
		decryptFunc := func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			encryptedData, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error(
					"failed to read request body", zap.Error(err),
				)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			data, err := decrypter.DecryptMsg(encryptedData)
			if err != nil {
				logger.Log.Info(
					"failed to decrypt request data", zap.Error(err),
				)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			readCloser := &decryptReadCloser{
				Reader: bytes.NewReader(data),
				rc:     r.Body,
			}
			r.Body = readCloser
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(decryptFunc)
	}
	return middlewareFunc
}

type decryptReadCloser struct {
	*bytes.Reader
	rc io.ReadCloser
}

func (drc *decryptReadCloser) Close() error {
	return drc.rc.Close()
}
