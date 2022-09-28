package transports

import (
	"os"
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/logz"
)

type FileWriter struct {
	name    string
	outfile *os.File
	write   logz.WriteFunc
	stream  chan logz.ThreadSaveEntry
}

func (f *FileWriter) Name() string {
	return f.name
}

func (f *FileWriter) Stream() chan logz.ThreadSaveEntry {
	return f.stream
}

func (f *FileWriter) Write(entry logz.Entry) error {
	return f.write(entry)
}

func (f *FileWriter) Cleanup() error {
	return f.outfile.Close()
}

// DefaultFileWriter uses FileWriteObserver to stream logs to an log file
func DefaultFileWriter(path string) logz.Builder {
	return func() (logz.Transport, error) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}

		outfile, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		return &FileWriter{
			name:    path,
			outfile: outfile,
			stream:  make(chan logz.ThreadSaveEntry, 1000),
			write:   NewIOWriteFunc(outfile, outfile, logz.DefaultFormatter),
		}, nil
	}
}

// SuggestedLogFileWriter creates a file writer that uses a suggested log file by tool and filename
func SuggestedLogFileWriter(tool, filename string) logz.Builder {
	return func() (logz.Transport, error) {
		logfile, err := logz.SuggestedFilePath(tool, filename)
		if err != nil {
			return nil, err
		}

		return DefaultFileWriter(logfile)()
	}
}
