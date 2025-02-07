package middleware

import (
	"net/http"
	"slices"
)

func AllowJSON(next http.Handler) http.Handler {
	jsonFunc := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Values(ContentType)
		if !slices.Contains(contentType, JSONMediaType) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(jsonFunc)
}
