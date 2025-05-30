package middleware

import (
	"net/http"
	"slices"

	"github.com/niksmo/runlytics/internal/server/app/http/header"
	"github.com/niksmo/runlytics/internal/server/app/http/mime"
)

func AllowJSON(next http.Handler) http.Handler {
	jsonFunc := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Values(header.ContentType)
		if !slices.Contains(contentType, mime.JSON) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(jsonFunc)
}
