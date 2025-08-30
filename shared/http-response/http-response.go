package httpresponse

import (
	"encoding/json"
	"net/http"

	sharedLogger "shared/logger"
)

// Response represents a response message
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// WriteResponse writes a JSON response with the given status code
func WriteResponse(w http.ResponseWriter, r *http.Request, serviceName string, response Response) {
	logger := sharedLogger.GetRequestLogger(r, serviceName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
