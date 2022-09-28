package cqrs

import (
	"fmt"

	"github.com/pkg/errors"
)

// RetryableError ...
type RetryableError struct {
	cause error
}

func (r RetryableError) Error() string {
	return fmt.Sprintf("a retryable error occurred: %s", r.cause.Error())
}
func (r RetryableError) Unwrap() error {
	return r.cause
}

func (r RetryableError) Cause() error {
	return r.cause
}

func (r RetryableError) Wrap(err error) error {
	r.cause = err
	return r
}

func NewRetryableError(err error) RetryableError {
	return RetryableError{err}
}

var PublishTimeoutErr = errors.New("failed to publish message within timeout")
