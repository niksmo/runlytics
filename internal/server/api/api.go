package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

const (
	ContentType     = "Content-Type"
	ContentEncoding = "Content-Encoding"
	AcceptEncoding  = "Accept-Encoding"
	JSONMediaType   = "application/json"
)

func debugLogRegister(endpoint string) {
	logger.Log.Debug("Register", zap.String("endpoint", endpoint))
}

func decodeJSON(r *http.Request, scheme any) error {
	if err := json.NewDecoder(r.Body).Decode(scheme); err != nil {
		errText := "Read body"
		logger.Log.Debug(errText, zap.Error(err))
		return fmt.Errorf("%s error: %w", errText, err)
	}
	return nil
}

func writeJSONResponse(w http.ResponseWriter, scheme any) {
	w.Header().Set(ContentType, JSONMediaType)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(scheme); err != nil {
		http.Error(w, server.ErrInternal.Error(), http.StatusInternalServerError)
		logger.Log.Debug("Encode JSON", zap.Error(err))
		return
	}
}
