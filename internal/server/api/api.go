package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

const (
	ContentType   = "Content-Type"
	JSONMediaType = "application/json"
)

func debugLogRegister(endpoint string) {
	logger.Log.Debug("Register", zap.String("endpoint", endpoint))
}

func verifyContentType(
	r *http.Request,
	mediaType string,
) error {
	contentType, ok := r.Header[ContentType]
	if !(ok && slices.Contains(contentType, mediaType)) {
		errText := fmt.Sprintf("expect %s content", mediaType)
		logger.Log.Debug(
			"Unsupported request Content-Type",
			zap.String(ContentType, strings.Join(contentType, "; ")),
			zap.String("Expected", mediaType),
		)
		return errors.New(errText)
	}
	return nil
}

func decodeJSON(
	r *http.Request,
	scheme any,
) error {
	if err := json.NewDecoder(r.Body).Decode(scheme); err != nil {
		errText := "incoming JSON object is not valid"
		logger.Log.Debug(errText, zap.Error(err))
		return errors.New(errText)
	}

	return nil
}

func writeJSONResponse(
	w http.ResponseWriter,
	scheme any,
) {
	w.Header().Set(ContentType, JSONMediaType)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(scheme); err != nil {
		errText := "Outgoing JSON object is not valid"
		http.Error(w, errText, http.StatusInternalServerError)
		logger.Log.Debug(
			errText,
			zap.Error(err),
		)
		return
	}
}

func writeTextErrorResponse(
	w http.ResponseWriter,
	statusCode int,
	text string,
) {
	http.Error(w, fmt.Sprintf("Error: %s", text), statusCode)
}
