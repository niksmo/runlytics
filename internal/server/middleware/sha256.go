package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

func VerifyAndWriteSHA256(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mdw := func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				logger.Log.Debug("Key is not using, skip header check")
				next.ServeHTTP(w, r)
				return
			}

			reqSHA256Hex := r.Header.Get("HashSHA256")
			if reqSHA256Hex == "" {
				logger.Log.Error("HashSHA256 header is empty or not exists")
				http.Error(
					w,
					"Require header 'HashSHA256'",
					http.StatusBadRequest,
				)
				return
			}
			logger.Log.Debug("Got hash", zap.String("hash", reqSHA256Hex))
			reqSHA256, err := hex.DecodeString(reqSHA256Hex)
			if err != nil {
				http.Error(
					w,
					"Require header 'HashSHA256' in hex format",
					http.StatusBadRequest,
				)
				return
			}

			h := hmac.New(sha256.New, []byte(key))
			_, err = io.Copy(h, r.Body)
			if err != nil {
				logger.Log.Error("Copy body to Hash", zap.Error(err))
				http.Error(
					w,
					server.ErrInternal.Error(),
					http.StatusInternalServerError,
				)
				return
			}
			if !hmac.Equal(h.Sum(nil), reqSHA256) {
				logger.Log.Error("Hash is not equal")
				http.Error(
					w,
					"Not equal 'HashSHA256'",
					http.StatusBadRequest,
				)
				return
			}
			logger.Log.Debug("Hash is equal, call next")

			//TO DO WriterWrapper
			next.ServeHTTP(w, r)

		}

		return http.HandlerFunc(mdw)
	}
}
