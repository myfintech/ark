package mhttp

import (
	"context"
	"net/http"

	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/pkg/errors"
)

// ResponseFormatOptions configures how the response will look
type ResponseFormatOptions struct {
	Pretty bool
	Format string
}

// DefaultResponseOptions preferred defaults for the response options
func DefaultResponseOptions() ResponseFormatOptions {
	return ResponseFormatOptions{
		Pretty: true,
		Format: "json",
	}
}

// XRequest the MANTL HTTP Framework XRequest
type XRequest struct {
	RawReq *http.Request
}

type XResponse struct {
	RawRes                http.ResponseWriter
	Errors                []HTTPError
	ResponseFormatOptions ResponseFormatOptions

	// Used for logging interceptors
	ResponseStatus int
	ResponseLength int
}

// AddError adds an error to the response
func (c *XResponse) AddError(err HTTPError) *XResponse {
	c.Errors = append(c.Errors, err)
	return c
}

// NewError creates an error, adds it to the response and returns its pointer
func (c XResponse) NewError(status int, code, message string) HTTPError {
	httpErr := NewHTTPError(status, code, message)
	c.AddError(httpErr)
	return httpErr
}

// HasErrors checks if XRequest.Errors length is greater than 0
func (c XResponse) HasErrors() bool {
	return len(c.Errors) > 0
}

// HighestErrorStatusCode iterates over list of response.Errors and returns the highest status code.
// If no errors are present we default to 500
func (c XResponse) HighestErrorStatusCode() int {
	if !c.HasErrors() {
		return http.StatusInternalServerError
	}
	var status int
	for _, err := range c.Errors {
		if err.Status > status {
			status = err.Status
		}
	}
	return status
}

// Fail sends a JSON response with the contents of response.Errors
// If no errors have been reported using response.AddError a default internal server error will be sent
func (c XResponse) Fail() error {
	if !c.HasErrors() {
		c.AddError(InternalServerError)
	}
	return c.Status(c.HighestErrorStatusCode()).JSON(&HTTPErrorResponse{
		Errors: c.Errors,
	})
}

// Write writes bytes to the response
func (c *XResponse) Write(data []byte) (int, error) {
	c.ResponseLength = len(data)
	return c.RawRes.Write(data)
}

// Status sets the response status on the request
func (c *XResponse) Status(status int) *XResponse {
	c.ResponseStatus = status
	return c
}

// Send responds to a request with a statusCode and plain text response
func (c XResponse) Send(message string) (int, error) {
	c.RawRes.WriteHeader(c.ResponseStatus)
	return c.RawRes.Write([]byte(message))
}

// SetContentType sets the Content-Type header
func (c *XResponse) SetContentType(contentType string) *XResponse {
	c.RawRes.Header().Set("Content-Type", contentType)
	return c
}

// JSON attempts to marshal data as JSON and send an HTTP response
// if marshalling fails then a JSONMarshal error is returned
func (c XResponse) JSON(data interface{}) (err error) {
	jsonData, err := utils.MarshalJSON(data, c.ResponseFormatOptions.Pretty)

	if err != nil {
		c.AddError(JSONMarshalError.WithContext(ErrorContext{
			"marshal_error": err.Error(),
		}))
		if sendError := c.Fail(); sendError != nil {
			err = errors.Wrap(sendError, "failed to send response")
		}
		return err
	}

	c.SetContentType("application/json")

	if _, sendErr := c.Send(jsonData); sendErr != nil {
		err = errors.Wrap(sendErr, "failed to send response")
	}
	return err
}

// AddErrorAndFail adds an error and fails the request immediately
func (c XResponse) AddErrorAndFail(httpError HTTPError) error {
	c.AddError(httpError)
	return c.Fail()
}

// NewErrorAndFail adds an error and fails the request immediately
func (c XResponse) NewErrorAndFail(status int, code, message string) error {
	c.NewError(status, code, message)
	return c.Fail()
}

// NewXResponse creates a new response struct with a writer
func NewXResponse(w http.ResponseWriter) *XResponse {
	return &XResponse{
		RawRes:                w,
		Errors:                make([]HTTPError, 0),
		ResponseFormatOptions: DefaultResponseOptions(),
		ResponseStatus:        200,
		ResponseLength:        0,
	}
}

// NewXRequest creates a new response struct with a writer
func NewXRequest(r *http.Request) *XRequest {
	return &XRequest{
		RawReq: r,
	}
}

// Next pass to the next handler
type Next func()

// XHandler a handler
type XHandler func(req *XRequest, res *XResponse, next Next)

func (xh XHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	xh(NewXRequest(r), NewXResponse(w), func() {})
}

// XHandlerFunc a compatibility interface with stdlib

// XMiddleware build a middleware chain
func XMiddleware(xh XHandler) RawMiddleware {
	return func(stdlibHandler http.Handler) http.Handler {
		return XHandler(func(req *XRequest, res *XResponse, next Next) {
			xh(req, res, func() {
				stdlibHandler.ServeHTTP(res.RawRes, req.RawReq)
			})
		})
	}
}

// AddValueToContext adds a value to the request context
func (c *XRequest) AddValueToContext(key, value interface{}) {
	c.RawReq = c.RawReq.WithContext(context.WithValue(c.RawReq.Context(), key, value))
}
