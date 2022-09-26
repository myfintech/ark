package logz

import (
	"io"

	"github.com/stretchr/testify/mock"
)

// StdLogger a standard interface defining methods for application logging
type StdLogger interface {
	Error(opts ...interface{})
	Errorf(format string, opts ...interface{})

	Warn(opts ...interface{})
	Warnf(format string, opts ...interface{})

	Info(opts ...interface{})
	Infof(format string, opts ...interface{})

	Debug(opts ...interface{})
	Debugf(format string, opts ...interface{})

	Trace(opts ...interface{})
	Tracef(format string, opts ...interface{})
}

// FieldLogger an extension ot StdLogger that appends contextual field data
type FieldLogger interface {
	StdLogger
	io.Writer
	Child(opts ...Option) FieldLogger
	Extend(opts ...Option) FieldLogger
	WithFields(Fields) FieldLogger
}

// Formatter is an interface that is expected to receive an Entry and serialize it into a []byte
type Formatter func(entry Entry) ([]byte, error)

// WriteFunc is a function passed to the Writer.RegisterTransport function.
// This method is used in a goroutine and is passed each log entry as they arrive.
// The behavior of the Writer is to attempt to deliver messages to each WriteFunc as fast as possible.
// If a WriteFunc cannot keep up and backpressure on the queue reaches 1000 messages will be dropped from this
// WriteFunc stream until it can handle additional capacity
type WriteFunc func(Entry) error

// NoOpLogger for testing purposes
type NoOpLogger struct{}

func (m NoOpLogger) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m NoOpLogger) Error(_ ...interface{})            {}
func (m NoOpLogger) Errorf(_ string, _ ...interface{}) {}
func (m NoOpLogger) Warn(_ ...interface{})             {}
func (m NoOpLogger) Warnf(_ string, _ ...interface{})  {}
func (m NoOpLogger) Info(_ ...interface{})             {}
func (m NoOpLogger) Infof(_ string, _ ...interface{})  {}
func (m NoOpLogger) Debug(_ ...interface{})            {}
func (m NoOpLogger) Debugf(_ string, _ ...interface{}) {}
func (m NoOpLogger) Trace(_ ...interface{})            {}
func (m NoOpLogger) Tracef(_ string, _ ...interface{}) {}
func (m NoOpLogger) Child(_ ...Option) FieldLogger     { return m }
func (m NoOpLogger) Extend(_ ...Option) FieldLogger    { return m }
func (m NoOpLogger) WithFields(Fields) FieldLogger     { return m }

// MockLogger used for test
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *MockLogger) Error(opts ...interface{}) {
	m.Called(opts...)
}

func (m *MockLogger) Errorf(format string, opts ...interface{}) {
	m.Called(format, opts)
}
func (m *MockLogger) Warn(opts ...interface{}) {
	m.Called(opts...)
}

func (m *MockLogger) Warnf(format string, opts ...interface{}) {
	m.Called(format, opts)
}

func (m *MockLogger) Info(opts ...interface{}) {
	m.Called(opts...)
}

func (m *MockLogger) Infof(format string, opts ...interface{}) {
	m.Called(format, opts)
}

func (m *MockLogger) Debug(opts ...interface{}) {
	m.Called(opts...)
}

func (m *MockLogger) Debugf(format string, opts ...interface{}) {
	m.Called(format, opts)
}

func (m *MockLogger) Trace(opts ...interface{}) {
	m.Called(opts...)
}

func (m *MockLogger) Tracef(format string, opts ...interface{}) {
	m.Called(format, opts)
}

func (m *MockLogger) Child(opts ...Option) FieldLogger {
	m.Called(opts)
	return m
}

func (m *MockLogger) Extend(opts ...Option) FieldLogger {
	m.Called(opts)
	return m
}

func (m *MockLogger) WithFields(f Fields) FieldLogger {
	m.Called(f)
	return m
}

// Injector an interface used to inject a field logger
type Injector interface {
	UseLogger(logger FieldLogger)
}

// Transport writes an entry to a given destination
type Transport interface {
	Name() string
	Stream() chan ThreadSaveEntry
	Write(entry Entry) error
	Cleanup() error
}

// Builder initializes a transport with a context
// If a builder returns an error its assumed that the log destination cannot be used and
// The error will be surfaced to the caller
type Builder func() (transport Transport, err error)
