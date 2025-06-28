package models

import "time"

// CommonErrorResponse for API errors
type CommonErrorResponse struct {
	Success   bool                `json:"success"`
	Timestamp time.Time           `json:"timestamp"`
	Message   string              `json:"message"`
	Errors    map[string][]string `json:"errors,omitempty"` // Validation errors
	Code      int                 `json:"code"`
	Method    string              `json:"method,omitempty"`
	Path      string              `json:"path,omitempty"`
}

// GenericSuccessResponse provides a flexible structure for all successful API responses.
// The 'Data' field can hold any Go struct that represents the actual payload.
type GenericSuccessResponse struct {
	Code    int         `json:"code"`            // HTTP Status Code (or custom code like 200000)
	Success bool        `json:"success"`         // Always true for success
	Data    interface{} `json:"data"`            // This is the flexible part! It can be any struct.
	Count   int         `json:"count,omitempty"` // Optional: useful for list responses
}
