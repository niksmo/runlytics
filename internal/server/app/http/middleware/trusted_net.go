package middleware

import (
	"net"
	"net/http"
)

func TrustedNet(ipNet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return trustedNetHandler{ipNet: ipNet, n: next}
	}
}

type trustedNetHandler struct {
	ipNet *net.IPNet
	n     http.Handler
}

func (h trustedNetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP := net.ParseIP(r.Header.Get(XRealIP))
	if clientIP == nil || !h.ipNet.Contains(clientIP) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	h.n.ServeHTTP(w, r)
}
