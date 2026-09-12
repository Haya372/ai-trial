package response

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body errorBody) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Code: code, Message: message})
}

func WriteValidationError(w http.ResponseWriter, details []ErrorDetail) {
	writeJSON(w, http.StatusBadRequest, errorBody{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: details,
	})
}
