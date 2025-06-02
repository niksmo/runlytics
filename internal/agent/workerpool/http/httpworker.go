package httpworker

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/agent/workerpool"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

const headerHashKey = "HashSHA256"

var (
	bufferPool     = sync.Pool{}
	gzipWriterPool = sync.Pool{}
)

func SendMetrics(
	ctx context.Context,
	m metrics.MetricsList,
	enc di.Encrypter,
	url, hk, ip string,
) error {
	const op = "httpworker.SendMetrics"
	log := logger.Log.With(
		zap.String("op", op), zap.String("url", url), zap.String("ip", ip),
	)
	buf, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
	} else {
		buf.Reset()
	}
	defer bufferPool.Put(buf)

	var sha256 string
	if err := makeReqData(buf, &sha256, m, hk, enc); err != nil {
		log.Fatal("failed to make request data", zap.Error(err))
	}

	req, err := newRequest(url, buf, sha256, ip)
	if err != nil {
		log.Fatal("failed to create request", zap.Error(err))
	}

	reqStart := time.Now()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn(
			"failed to do request",
			zap.Duration("resTime", time.Since(reqStart)), zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}
	defer res.Body.Close()
	log.Info(
		"got response",
		zap.String("status", res.Status),
		zap.Duration("resTime", time.Since(reqStart)),
	)

	if _, err := readResData(res); err != nil {
		log.Warn("failed to read response data", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func makeReqData(
	buf *bytes.Buffer,
	sha256 *string,
	metrics []metrics.Metrics,
	key string,
	encrypter di.Encrypter,
) error {
	const op = "httpworker.makeReqData"

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if key != "" {
		*sha256, err = workerpool.GetHashString(jsonData, key)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

	}

	gzipWriter, ok := gzipWriterPool.Get().(*gzip.Writer)
	if !ok {
		gzipWriter = gzip.NewWriter(buf)
	} else {
		gzipWriter.Reset(buf)
	}
	defer gzipWriterPool.Put(gzipWriter)

	if _, err = gzipWriter.Write(jsonData); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	encryptedData, err := encrypter.EncryptMsg(buf.Bytes())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	buf.Reset()
	buf.Write(encryptedData)

	return nil
}

func newRequest(
	URL string, body *bytes.Buffer, sha256 string, outboundIP string,
) (*http.Request, error) {
	const op = "httpworker.newRequest"

	request, err := http.NewRequest("POST", URL, body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")
	if sha256 != "" {
		request.Header.Set(headerHashKey, sha256)
	}
	if outboundIP != "" {
		request.Header.Set("X-Real-IP", outboundIP)
	}
	return request, nil
}

func readResData(res *http.Response) ([]byte, error) {
	const op = "httpworker.readResData"

	if res.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		res.Body = gzipReader
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return data, nil
}
