package middleware

import (
	"net/http"
	"strings"

	"github.com/niksmo/runlytics/pkg/httpserver/header"
)

func AllowContentEncoding(
	contentEncoding ...string,
) func(next http.Handler) http.Handler {
	allowedEncodings := make(map[string]struct{}, len(contentEncoding))
	for _, encoding := range contentEncoding {
		encoding = strings.ToLower(strings.TrimSpace(encoding))
		allowedEncodings[encoding] = struct{}{}
	}

	mdw := func(next http.Handler) http.Handler {
		acceptFunc := func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			reqEncodings := r.Header.Values(header.ContentEncoding)

			for _, reqEncoding := range reqEncodings {
				reqEncoding = strings.ToLower(strings.TrimSpace(reqEncoding))
				if _, ok := allowedEncodings[reqEncoding]; !ok {
					w.WriteHeader(http.StatusUnsupportedMediaType)
					return
				}
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(acceptFunc)
	}

	return mdw
}
