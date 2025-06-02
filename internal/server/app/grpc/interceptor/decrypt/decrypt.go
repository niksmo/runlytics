package decrypt

import (
	"context"

	"github.com/niksmo/runlytics/pkg/di"
	"google.golang.org/grpc"
)

func New(decrypter di.Decrypter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		return handler(ctx, req)
	}
}
