package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

func Decrypt(decrypter di.Decrypter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &decryptHandler{d: decrypter, n: next}
	}
}

type decryptHandler struct {
	d di.Decrypter
	n http.Handler
}

func (h *decryptHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength == 0 {
		h.n.ServeHTTP(w, r)
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

	data, err := h.d.DecryptMsg(encryptedData)
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
	h.n.ServeHTTP(w, r)
}

type decryptReadCloser struct {
	*bytes.Reader
	rc io.ReadCloser
}

func (drc *decryptReadCloser) Close() error {
	return drc.rc.Close()
}
