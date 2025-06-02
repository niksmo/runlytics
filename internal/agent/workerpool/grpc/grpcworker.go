package grpcworker

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/niksmo/runlytics/internal/agent/workerpool"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	pb "github.com/niksmo/runlytics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func SendMetrics(
	ctx context.Context,
	m metrics.MetricsList,
	enc di.Encrypter,
	addr, hk, ip string,
) error {
	const op = "grpcworker.SendMetrics"
	log := logger.Log.With(
		zap.String("op", op), zap.String("addr", addr), zap.String("ip", ip),
	)
	c, err := getClient(addr)
	if err != nil {
		log.Fatal("failed to get grpc client", zap.Error(err))
	}
	req, err := newRequest(m)
	if err != nil {
		log.Fatal("failed crate new request", zap.Error(err))
	}

	md, err := newMetadata(req.GetMetrics(), hk, ip)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

	reqStart := time.Now()
	res, err := c.BatchUpdate(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info(
		"got response",
		zap.Duration("resTime", time.Since(reqStart)),
		zap.Uint32("updatedCount", res.GetUpdatedCount()),
	)

	return nil
}

func getClient(addr string) (pb.RunlyticsClient, error) {
	const op = "grpcworker.getClient"
	conn, err := grpc.NewClient(
		addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return pb.NewRunlyticsClient(conn), nil
}

func newRequest(m metrics.MetricsList) (*pb.BatchUpdateRequest, error) {
	const op = "grpcworker.newRequest"
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(m); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &pb.BatchUpdateRequest{Metrics: b.Bytes()}, nil
}

func newMetadata(out []byte, key, ip string) (metadata.MD, error) {
	const op = "grpcworker.newMetadata"
	md := metadata.New(map[string]string{})

	if key != "" {
		hash, err := workerpool.GetHashString(out, key)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		md.Append("HashSHA256", hash)
	}

	md.Append("X-Real-IP", ip)

	return md, nil
}
