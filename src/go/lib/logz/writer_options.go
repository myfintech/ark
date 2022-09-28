package logz

import (
	"fmt"

	"github.com/pkg/errors"
)

// Option is a function that applies mutations to a *Writer
type Option func(w *Writer) error

// WithLevel sets the level for this logger
func WithLevel(level Level) Option {
	return func(w *Writer) error {
		w.level = level
		return nil
	}
}

// WithParent sets the level for this logger
func WithParent(parent chan ThreadSaveEntry) Option {
	return func(w *Writer) error {
		w.parent = parent
		return nil
	}
}

// WithLevelString sets the level for this logger
func WithLevelString(levelStr string) Option {
	return func(w *Writer) error {
		level, err := ParseLevel(levelStr)
		if err != nil {
			return err
		}
		w.level = level
		return nil
	}
}

// WithFields an option that applies fields to a *Writer
func WithFields(fields Fields) Option {
	return func(w *Writer) error {
		if w.fields == nil {
			w.fields = make([][]string, 0)
		}
		dst := make([][]string, len(w.fields))
		for i, v := range w.fields {
			pair := make([]string, len(v))
			copy(pair, v)
			dst[i] = pair
		}
		for s, i := range fields {
			w.fields = append(dst, []string{s, fmt.Sprintf("%s", i)})
		}
		return nil
	}
}

// WithQueue an Option that applies a queue channel to a *Writer
func WithQueue(queue chan ThreadSaveEntry) Option {
	return func(w *Writer) error {
		w.queue = queue
		return nil
	}
}

// WithMux multiplexes log.Entry to all provided Observer
func WithMux(builders ...Builder) Option {
	return func(w *Writer) error {
		for _, builder := range builders {
			transport, err := builder()
			if err != nil {
				return errors.Wrap(err, "failed to build transport")
			}

			if err = w.RegisterTransport(transport); err != nil {
				return errors.Wrap(err, "failed to register transport")
			}
		}

		w.start()

		return nil
	}
}

// WithDisableAutoConsume turns off auto consumption of messages
// By default the logger will start multiplexing
// This should really only be used in tests that need to verify channel behavior
func WithDisableAutoConsume() Option {
	return func(w *Writer) error {
		w.disableAutoConsume = true
		return nil
	}
}
