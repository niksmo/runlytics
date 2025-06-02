package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"net/http"
	"sync"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

const headerHashKey = "HashSHA256"

var hashPool = sync.Pool{}

var (
	ErrNotEqualHash = errors.New("invalid sha256 hash sum")
)

func VerifyAndWriteSHA256(key string, method ...string) func(http.Handler) http.Handler {
	verifyMethods := map[string]struct{}{}
	for _, m := range method {
		verifyMethods[m] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		mdw := func(w http.ResponseWriter, r *http.Request) {
			if key == "" || r.Header.Get(headerHashKey) == "" {
				logger.Log.Debug("Key is not using, skip header check")
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := verifyMethods[r.Method]; !ok {
				logger.Log.Debug("Not verifying method, skip header check")
				next.ServeHTTP(w, r)
				return
			}

			reqSHA256Hex := r.Header.Get(headerHashKey)

			reqSHA256, err := hex.DecodeString(reqSHA256Hex)
			if err != nil {
				logger.Log.Debug("Decode hex hash", zap.Error(err))
				http.Error(
					w,
					"Require header 'HashSHA256' in hex format",
					http.StatusBadRequest,
				)
				return
			}

			r.Body = newHashReader(r.Body, key, reqSHA256)
			w = newHashWriter(w, key)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(mdw)
	}
}

type hashReader struct {
	io.ReadCloser
	key     []byte
	compare []byte
	buf     bytes.Buffer
}

func newHashReader(
	wrapped io.ReadCloser, key string, comparedHash []byte,
) *hashReader {
	return &hashReader{
		ReadCloser: wrapped,
		key:        []byte(key),
		compare:    comparedHash,
	}
}

func (hashReader *hashReader) Read(p []byte) (int, error) {
	n, err := hashReader.ReadCloser.Read(p)
	if err == nil {
		hashReader.buf.Write(p[:n])
	}

	if errors.Is(err, io.EOF) {
		hashReader.buf.Write(p[:n])

		h, ok := hashPool.Get().(hash.Hash)
		if !ok {
			h = hmac.New(sha256.New, hashReader.key)
		}
		defer hashPool.Put(h)
		h.Reset()

		if _, hErr := h.Write(hashReader.buf.Bytes()); hErr != nil {
			return 0, hErr
		}

		if !hmac.Equal(h.Sum(nil), hashReader.compare) {
			logger.Log.Debug("Hash is not equal")
			return 0, ErrNotEqualHash
		}
	}

	return n, err
}

type hashWriter struct {
	http.ResponseWriter
	hash hash.Hash
}

func newHashWriter(wrapped http.ResponseWriter, key string) *hashWriter {
	return &hashWriter{
		ResponseWriter: wrapped,
		hash:           hmac.New(sha256.New, []byte(key)),
	}
}

func (hashWriter *hashWriter) Write(p []byte) (int, error) {
	n, err := hashWriter.hash.Write(p)
	if err != nil {
		return n, err
	}
	n, err = hashWriter.ResponseWriter.Write(p)
	return n, err
}

func (hashWriter *hashWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		hashWriter.Header().Set(
			headerHashKey, hex.EncodeToString(hashWriter.hash.Sum(nil)),
		)
	}
	hashWriter.ResponseWriter.WriteHeader(statusCode)
}
