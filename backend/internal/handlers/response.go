package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Status  int    `json:"status"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondJSON writes a JSON response with status code and data.
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// RespondSuccess writes a standard successful JSON response.
func RespondSuccess(w http.ResponseWriter, status int, data interface{}, message string) {
	RespondJSON(w, status, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// RespondError writes a standard error JSON response.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondJSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    code,
		Status:  status,
	})
}
