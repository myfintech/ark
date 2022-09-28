package httpresputils

import (
	"net/http"

	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/pkg/errors"
)

// Response is a convenient wrapper for http.ResponseWriter
type Response struct {
	Writer http.ResponseWriter
	Pretty bool
	Errors []*HTTPError
}

// AddError adds an HTTPError to the Response.Errors list
func (r *Response) AddError(err *HTTPError) {
	r.Errors = append(r.Errors, err)
}

// NewError creates an error, adds it to the response and returns its pointer
func (r *Response) NewError(status int, code, message string) *HTTPError {
	httpErr := NewHTTPError(status, code, message)
	r.AddError(httpErr)
	return httpErr
}

// HasErrors checks if response.Errors length is greater than 0
func (r *Response) HasErrors() bool {
	return len(r.Errors) > 0
}

// HighestErrorStatusCode iterates over list of response.Errors and returns the highest status code
// If no errors are present a 500 status is returned
func (r *Response) HighestErrorStatusCode() int {
	if !r.HasErrors() {
		return http.StatusInternalServerError
	}
	var status int
	for _, err := range r.Errors {
		if err.Status > status {
			status = err.Status
		}
	}
	return status
}

// Fail sends a JSON response with the contents of response.Errors
// If no errors have been reported using response.AddError a default internal server error will be sent
func (r *Response) Fail() error {
	if !r.HasErrors() {
		r.AddError(&InternalServerError)
	}
	return r.SendJSON(r.HighestErrorStatusCode(), &HTTPErrorResponse{
		Errors: r.Errors,
	})
}

// Send responds to a request with a statusCode and plain text response
func (r *Response) Send(status int, message string) (int, error) {
	r.Writer.WriteHeader(status)
	return r.Writer.Write([]byte(message))
}

// SetContentType sets the Content-Type header
func (r *Response) SetContentType(contentType string) {
	r.Writer.Header().Set("Content-Type", contentType)
}

// SendJSON attempts to marshal data as JSON and send an HTTP response
// if marshalling fails then a JSONMarshal error is returned
func (r *Response) SendJSON(status int, data interface{}) error {
	jsonData, err := utils.MarshalJSON(data, r.Pretty)

	if err != nil {
		err = errors.Wrap(err, "Failed to marshal JSON")
		r.AddError(JSONMarshalError.WithContext(ErrorContext{
			"marshal_error": err.Error(),
		}))
		if sendError := r.Fail(); sendError != nil {
			err = errors.Wrap(sendError, "failed to send response")
		}
		return err
	}

	r.SetContentType("application/json")

	if _, sendErr := r.Send(status, jsonData); sendErr != nil {
		err = errors.Wrap(sendErr, "failed to send response")
	}
	return err
}

// AddErrorAndFail adds an error and fails the request immediately
func (r *Response) AddErrorAndFail(httpError *HTTPError) error {
	r.AddError(httpError)
	return r.Fail()
}

// NewErrorAndFail adds an error and fails the request immediately
func (r *Response) NewErrorAndFail(status int, code, message string) error {
	r.NewError(status, code, message)
	return r.Fail()
}

// NewResponse creates a new response struct with a writer
func NewResponse(writer http.ResponseWriter) *Response {
	return &Response{
		Writer: writer,
		Pretty: true,
		Errors: make([]*HTTPError, 0),
	}
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(status int, code, message string) *HTTPError {
	return &HTTPError{
		Status:        status,
		Code:          code,
		Message:       message,
		Context:       ErrorContext{},
		SecureContext: SecureContext{},
	}
}
