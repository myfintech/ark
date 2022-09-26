package mhttp

import (
	"net/http"

	"github.com/myfintech/ark/src/go/lib/utils"
)

// HTTPError an object that represents an error that can be sent to the client
type HTTPError struct {
	Status  int          `json:"status"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Context ErrorContext `json:"context"`

	// TODO: Implement secure_context (experimental)
	SecureContext SecureContext `json:"secure_context"`
}

// ErrorContext is a map[string]interface{}. Hopefully, JSON serializable.
type ErrorContext map[string]interface{}

// SecureContext is a map[string]interface{}. Hopefully, JSON serializable.
type SecureContext map[string]interface{}

// WithContext returns a new copy of the HTTPError with context
func (h HTTPError) WithContext(ctx ErrorContext) HTTPError {
	return HTTPError{
		Status:  h.Status,
		Code:    h.Code,
		Message: h.Message,
		Context: utils.MergeMaps(h.Context, ctx),
	}
}

// WithSecureContext returns a new copy of the HTTP error with a SecureContext
func (h HTTPError) WithSecureContext(ctx SecureContext) HTTPError {
	return HTTPError{
		Status:        h.Status,
		Code:          h.Code,
		Message:       h.Message,
		Context:       h.Context,
		SecureContext: utils.MergeMaps(h.SecureContext, ctx),
	}
}

// HTTPErrorResponse
type HTTPErrorResponse struct {
	Errors []HTTPError `json:"errors"`
}

// JSONMarshalError an error used to indicate a JSON marshalling error
var JSONMarshalError = HTTPError{
	Code:    "JSON_MARSHAL_ERROR",
	Message: "Failed to marshal JSON",
	Status:  http.StatusInternalServerError,
}

// InternalServerError an error used to indicate an internal server error
var InternalServerError = HTTPError{
	Status:  http.StatusInternalServerError,
	Code:    "INTERNAL_SERVER_ERROR",
	Message: "Internal Server Error",
	Context: ErrorContext{},
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(status int, code, message string) HTTPError {
	return HTTPError{
		Status:        status,
		Code:          code,
		Message:       message,
		Context:       ErrorContext{},
		SecureContext: SecureContext{},
	}
}
