package netcheck

import (
	"context"
	"net"

	"github.com/niksmo/runlytics/internal/server/app/grpc/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const xRealIP = "X-Real-IP"

var ErrNotAllowedIP = status.Error(
	codes.Unauthenticated, "not allowed zone",
)

func New(ipNet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if ipNet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, interceptor.ErrMissingMetadata
		}

		if !validMD(md.Get(xRealIP)) {
			return nil, interceptor.ErrMissingMetadata
		}

		clientIP := md.Get(xRealIP)[0]
		if !validIPNet(clientIP, ipNet) {
			return nil, ErrNotAllowedIP
		}

		return handler(ctx, req)
	}
}

func validMD(ip []string) bool {
	return len(ip) >= 1
}

func validIPNet(cIP string, ipNet *net.IPNet) bool {
	ip := net.ParseIP(cIP)
	return ipNet.Contains(ip)

}
