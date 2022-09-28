package transports

import (
	"io"
	"os"

	"github.com/myfintech/ark/src/go/lib/logz"
)

type IOWriter struct {
	name   string
	write  logz.WriteFunc
	stream chan logz.ThreadSaveEntry
}

func (i *IOWriter) Stream() chan logz.ThreadSaveEntry {
	return i.stream
}

func (i *IOWriter) Name() string {
	return i.name
}

func (i *IOWriter) Write(entry logz.Entry) error {
	return i.write(entry)
}

func (i *IOWriter) Cleanup() error {
	return nil
}

// NewIOWriteFunc a transport function that accepts a log.Entry and writes it to the provided io.Writers
func NewIOWriteFunc(stdOut, stdErr io.Writer, formatter logz.Formatter) logz.WriteFunc {
	return func(entry logz.Entry) (err error) {
		msg, err := formatter(entry)
		if err != nil {
			return
		}

		var out io.Writer
		if entry.Level == logz.InfoLevel {
			out = stdOut
		} else {
			out = stdErr
		}

		if _, err = out.Write(msg); err != nil {
			return
		}
		return
	}
}

// DefaultIOWriter uses IOWriteObserver to stream logs to os.Stdout and os.Stderr using logz.DefaultFormatter
func DefaultIOWriter() (logz.Transport, error) {
	return &IOWriter{
		name:   "default",
		stream: make(chan logz.ThreadSaveEntry, 1000),
		write:  NewIOWriteFunc(os.Stdout, os.Stderr, logz.DefaultFormatter),
	}, nil
}

// JsonIOLogWriter uses IOWriteObserver to stream logs to os.Stdout and os.Stderr using logz.JsonFormatter
