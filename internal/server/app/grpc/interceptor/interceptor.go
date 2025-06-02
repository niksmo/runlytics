package interceptor

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrMissingMetadata = status.Error(
		codes.InvalidArgument, "missing metadata",
	)

	ErrInvalidRequestMessage = status.Error(
		codes.InvalidArgument, "invalid request message type",
	)
)
