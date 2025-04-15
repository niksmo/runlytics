package jsonhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ReadRequest(r *http.Request, scheme any) error {
	if err := json.NewDecoder(r.Body).Decode(scheme); err != nil {
		return fmt.Errorf("decode request body error: %w", err)
	}
	return nil
}

func WriteResponse(w http.ResponseWriter, statusCode int, scheme any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(scheme); err != nil {
		return fmt.Errorf("encode response body error: %w", err)
	}
	return nil
}
