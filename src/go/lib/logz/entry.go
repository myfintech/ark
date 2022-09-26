package logz

import (
	"time"
)

type Entry struct {
	// Contains all the fields set by the user.
	Fields Fields `json:"data,omitempty"`

	// Timestamp at which the log entry was created
	Timestamp time.Time `json:"timestamp"`

	// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
	// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
	Level Level `json:"level"`

	// Calling method, with package name
	Caller string `json:"caller,omitempty"`

	// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
	Message string `json:"message"`

	Error error `json:"error,omitempty"`
}

type ThreadSaveEntry struct {
	// Contains all the fields set by the user.
	Fields [][]string `json:"data,omitempty"`

	// Timestamp at which the log entry was created
	Timestamp time.Time `json:"timestamp"`

	// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
	// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
	Level Level `json:"level"`

	// Calling method, with package name
	Caller string `json:"caller,omitempty"`

	// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
	Message string `json:"message"`

	Error error `json:"error,omitempty"`
}
