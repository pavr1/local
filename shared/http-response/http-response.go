package httpresponse

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Response represents a response message
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// WriteResponse writes a JSON response with the given status code
func WriteResponse(w http.ResponseWriter, logger *logrus.Logger, serviceName string, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
