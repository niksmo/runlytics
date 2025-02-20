package worker

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

const headerHashKey = "HashSHA256"

type JobErr struct {
	jobID int64
	err   error
}

func (e *JobErr) ID() int64 {
	return e.jobID
}

func (e *JobErr) Err() error {
	return e.err
}

func Run(
	jobCh <-chan di.Job,
	errCh chan<- di.JobErr,
	URL string,
	key string,
	HTTPClient *http.Client,
) {
	for job := range jobCh {
		logger.Log.Info("Start job", zap.Int64("jobID", job.ID()))
		body, sha256 := makeRequestBody(job.Payload(), key)
		start := time.Now()
		res, err := HTTPClient.Do(createRequest(URL, body, sha256))
		if err != nil {
			errCh <- &JobErr{jobID: job.ID(), err: err}
			logger.Log.Info(
				"Got response",
				zap.Int64("jobID", job.ID()),
				zap.String("URL", URL),
				zap.String("method", "POST"),
				zap.Duration("duration", time.Since(start)),
				zap.Error(err),
			)
			continue
		}

		data := readBody(res)
		logger.Log.Info(
			"Got response",
			zap.Int64("jobID", job.ID()),
			zap.String("URL", URL),
			zap.String("method", "POST"),
			zap.Duration("duration", time.Since(start)),
			zap.Int("statusCode", res.StatusCode),
			zap.String("hash", res.Header.Get(headerHashKey)),
			zap.String("data", string(data)),
		)
	}
}

func getHexHashSHA256(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(data)
	if err != nil {
		logger.Log.Panic("Write to Hash", zap.Error(err))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func makeRequestBody(
	metrics []metrics.MetricsUpdate, key string,
) (body io.Reader, hexSHA256 string) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Panic("Encode to JSON error", zap.Error(err))
	}

	if key != "" {
		hexSHA256 = getHexHashSHA256(jsonData, key)
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err = gzipWriter.Write(jsonData); err != nil {
		logger.Log.Panic("Write gzip", zap.Error(err))
	}
	gzipWriter.Close()

	body = &buf
	return
}

func createRequest(URL string, body io.Reader, sha256 string) *http.Request {
	request, err := http.NewRequest("POST", URL, body)
	if err != nil {
		logger.Log.Panic("Error while creating http request", zap.Error(err))
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")
	if sha256 != "" {
		request.Header.Set(headerHashKey, sha256)
	}
	return request
}

func readBody(response *http.Response) []byte {
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(response.Body)
		if err != nil {
			logger.Log.Panic("Error while creating new gzip reader", zap.Error(err))
		}
		response.Body = gzipReader
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Log.Error("Read response data", zap.Error(err))
	}
	return data

}
