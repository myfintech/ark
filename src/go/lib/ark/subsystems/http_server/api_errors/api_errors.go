package api_errors

import (
	"net/http"
)

// APIError the base element of an API error
type APIError struct {
	Message string                 `json:"message"`
	Status  int                    `json:"status"`
	Code    string                 `json:"code"`
	Context map[string]interface{} `json:"context"`
}

func (e APIError) Error() string {
	if err, ok := e.Context["error"]; ok {
		switch v := err.(type) {
		case string:
			return v
		case error:
			return v.Error()
		}
	}
	return e.Message
}

// WithContext creates a copy of the error with added context values
func (e APIError) WithContext(context map[string]interface{}) APIError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}

	for s, i := range context {
		e.Context[s] = i
	}
	return e
}

// WithField appends an arbitrary field and value to the context map
func (e APIError) WithField(field string, value interface{}) APIError {
	return e.WithContext(map[string]interface{}{
		field: value,
	})
}

// WithErr returns a new instance of the error with added error context values
func (e APIError) WithErr(err error) APIError {
	return e.WithContext(map[string]interface{}{
		"error": err.Error(),
	})
}

// InternalServerError a standard API error message
var InternalServerError = APIError{
	Message: "an internal server error occurred",
	Status:  http.StatusInternalServerError,
	Code:    "INTERNAL_SERVER_ERROR",
}

// NotFoundError an error indicating the resource couldn't be locating using the provided parameters
var NotFoundError = APIError{
	Message: "no resource found",
	Status:  http.StatusNotFound,
	Code:    "NOT_FOUND",
}

// BadRequest an error indicating a bad request
var BadRequest = APIError{
	Message: "bad request",
	Status:  http.StatusBadRequest,
	Code:    "BAD_REQUEST",
}
