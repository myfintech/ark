package logz

import "fmt"

// Error writes an error level message
func (w *Writer) Error(opts ...interface{}) {
	w.write(Entry{
		Level:   ErrorLevel,
		Message: fmt.Sprint(opts...),
	})
}

// Errorf writes an error level message with formatting
func (w *Writer) Errorf(format string, opts ...interface{}) {
	w.write(Entry{
		Level:   ErrorLevel,
		Message: fmt.Sprintf(format, opts...),
	})
}

// Warn writes a warn level message
func (w *Writer) Warn(opts ...interface{}) {
	w.write(Entry{
		Level:   WarnLevel,
		Message: fmt.Sprint(opts...),
	})
}

// Warnf writes a warn level message with formatting
func (w *Writer) Warnf(format string, opts ...interface{}) {
	w.write(Entry{
		Level:   WarnLevel,
		Message: fmt.Sprintf(format, opts...),
	})
}

// Info writes an info level message
func (w *Writer) Info(opts ...interface{}) {
	w.write(Entry{
		Level:   InfoLevel,
		Message: fmt.Sprint(opts...),
	})
}

// Infof writes an info level message with formatting
func (w *Writer) Infof(format string, opts ...interface{}) {
	w.write(Entry{
		Level:   InfoLevel,
		Message: fmt.Sprintf(format, opts...),
	})
}

// Debug writes a debug level message
func (w *Writer) Debug(opts ...interface{}) {
	w.write(Entry{
		Level:   DebugLevel,
		Message: fmt.Sprint(opts...),
	})
}

// Debugf writes a debug level message with formatting
func (w *Writer) Debugf(format string, opts ...interface{}) {
	w.write(Entry{
		Level:   DebugLevel,
		Message: fmt.Sprintf(format, opts...),
	})
}

// Trace writes a trace level message
func (w *Writer) Trace(opts ...interface{}) {
	w.write(Entry{
		Level:   TraceLevel,
		Message: fmt.Sprint(opts...),
	})
}

// Tracef writes a trace level message with formatting
func (w *Writer) Tracef(format string, opts ...interface{}) {
	w.write(Entry{
		Level:   TraceLevel,
		Message: fmt.Sprintf(format, opts...),
	})
}
