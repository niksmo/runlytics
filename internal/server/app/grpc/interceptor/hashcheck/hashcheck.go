package hashcheck

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"sync"

	"github.com/niksmo/runlytics/internal/server/app/grpc/interceptor"
	pb "github.com/niksmo/runlytics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const hashSHA256 = "HashSHA256"

var ErrInvalidHash = status.Error(codes.InvalidArgument, "invalid hash")

var hashPool = sync.Pool{}

func New(key string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if key == "" || validMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, interceptor.ErrMissingMetadata
		}

		if !validMD(md.Get(hashSHA256)) {
			return nil, interceptor.ErrMissingMetadata
		}

		typedReq, ok := req.(*pb.BatchUpdateRequest)
		if !ok {
			return nil, interceptor.ErrInvalidRequestMessage
		}

		hash := md.Get(hashSHA256)[0]
		if !validHash(key, hash, typedReq.GetMetrics()) {
			return nil, ErrInvalidHash
		}

		return handler(ctx, req)
	}
}

func validMethod(method string) bool {
	return method != pb.Runlytics_BatchUpdate_FullMethodName
}

func validMD(hash []string) bool {
	return len(hash) >= 1
}

func validHash(key, hashString string, in []byte) bool {
	hashSum, err := hex.DecodeString(hashString)
	if err != nil {
		return false
	}

	h, ok := hashPool.Get().(hash.Hash)
	if !ok {
		h = hmac.New(sha256.New, []byte(key))
	} else {
		h.Reset()
	}
	defer hashPool.Put(h)

	if _, err := h.Write(in); err != nil {
		return false
	}

	if !hmac.Equal(h.Sum(nil), hashSum) {
		return false
	}
	return true
}
