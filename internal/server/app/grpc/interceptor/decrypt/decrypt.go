package decrypt

import (
	"context"

	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrInvalidPayload = status.Error(
	codes.InvalidArgument, "invalid message payload",
)

func New(decrypter di.Decrypter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		switch r := req.(type) {
		case *proto.BatchUpdateRequest:
			data, err := decrypter.DecryptMsg(r.GetMetrics())
			if err != nil {
				return nil, ErrInvalidPayload
			}
			r.Metrics = data
			return handler(ctx, r)
		default:
			return handler(ctx, req)
		}

	}
}
