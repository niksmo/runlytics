package grpcapi

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	pb "github.com/niksmo/runlytics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedRunlyticsServer
	batchUpdateService di.BatchUpdateService
}

func Register(
	gRPCServer *grpc.Server,
	batchUpdateService di.BatchUpdateService,
) {
	pb.RegisterRunlyticsServer(
		gRPCServer, &serverAPI{batchUpdateService: batchUpdateService},
	)
}

func (s *serverAPI) BatchUpdate(
	ctx context.Context, in *pb.BatchUpdateRequest,
) (*pb.BatchUpdateResponse, error) {
	b := bytes.NewReader(in.GetMetrics())

	var ml metrics.MetricsList
	err := gob.NewDecoder(b).Decode(&ml)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, "failed to decode",
		)
	}

	err = ml.Verify(
		metrics.VerifyID,
		metrics.VerifyType,
		metrics.VerifyDelta,
		metrics.VerifyValue,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.batchUpdateService.BatchUpdate(ctx, ml)
	if err != nil {
		return nil, status.Error(
			codes.Internal, "failed to update",
		)
	}

	return &pb.BatchUpdateResponse{UpdatedCount: uint32(len(ml))}, nil
}
