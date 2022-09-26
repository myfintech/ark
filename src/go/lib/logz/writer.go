package logz

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/go-multierror"
)

// Writer implements the FieldLogger interface and provides a channel based write mechanism
// An Option can be provided that observes the queue and transports Entry records to a log sink
type Writer struct {
	fields             [][]string
	queue              chan ThreadSaveEntry
	parent             chan ThreadSaveEntry
	error              error
	level              Level
	ctx                context.Context
	wg                 *sync.WaitGroup
	eg                 *errgroup.Group
	once               *sync.Once
	transports         *sync.Map
	disableAutoConsume bool
}

func (w *Writer) start() {
	w.once.Do(func() {
		w.runPublisher()
	})
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.write(Entry{
		Level:   InfoLevel,
		Message: string(p),
	})
	return len(p), nil
}

// InitError check this property to validate that calling New did not result in a error from an option
func (w *Writer) InitError() error {
	return w.error
}

// Child returns a new logger instance that inherits from its parent
// By default it receives its parents fields and level
// All log messages will be propagated to the parent and written to its parents transports
// This new instance may receive its own transports which will not receive logs from the parent
// This logger may be closed independently of its parent
func (w *Writer) Child(opts ...Option) FieldLogger {
	fields := make(Fields)
	for _, v := range w.fields {
		fields[v[0]] = v[1]
	}

	opts = append([]Option{
		WithFields(fields),
		WithParent(w.queue),
		WithLevel(w.level),
	}, opts...)

	return New(w.ctx, opts...)
}

// WithFields is a lighter-weight alternative to Child
// The Child method is very heavy and creates a new instance with channels, transports, and the works.
// When used improperly it can cause memory leaks and the most common task is to create a logger with additional fields
func (w *Writer) WithFields(fields Fields) FieldLogger {
	return w.Extend(WithFields(fields))
}

// Extend is an lighter-weight alternative to Child that allows reuse of parent logger options
// This will not create transports but instead reuse the parents queue
func (w *Writer) Extend(opts ...Option) FieldLogger {
	wc := &Writer{
		level:      w.level,
		fields:     w.fields,
		queue:      w.queue,
		once:       w.once,
		wg:         w.wg,
		transports: w.transports,
	}

	for _, opt := range opts {
		if err := opt(w); err != nil {
			w.error = multierror.Append(w.error, err)
		}
	}

	return wc
}

func (w *Writer) write(entry Entry) {
	if w.level < entry.Level {
		return
	}

	if entry.Level > DebugLevel {
		entry.Caller = w.getCaller(3)
	}

	threadSaveEntry := ThreadSaveEntry{
		Fields:    w.fields,
		Timestamp: time.Now(),
		Level:     entry.Level,
		Caller:    entry.Caller,
		Message:   entry.Message,
		Error:     entry.Error,
	}

	w.writeOrDropWithWarning("self", w.queue, threadSaveEntry)
	w.writeOrDropWithWarning("parent", w.parent, threadSaveEntry)
}

func (w *Writer) getCaller(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprint(file, ":", line)
}

func (w *Writer) writeOrDropWithWarning(name string, dst chan ThreadSaveEntry, entry ThreadSaveEntry) {
	if dst == nil {
		return
	}

	select {
	case dst <- entry:
	default:
		caller := w.getCaller(4)
		fmt.Printf("[%s](%s.logger) is at capacity or closed (%d) \n", caller, name, len(dst))
	}
}

// Close stops new Entry records from being produced by the logger allowing the caller to Wait
// until all messages have been handled by all Observers
func (w *Writer) Close() {
	close(w.queue)
}

func (w *Writer) closeTransportStreams() {
	w.transports.Range(func(key, value interface{}) bool {
		transport := value.(Transport)
		close(transport.Stream())
		return true
	})
}

func (w *Writer) publishToAllTransportStreams(entry ThreadSaveEntry) {

	w.transports.Range(func(key, value interface{}) bool {

		transport := value.(Transport)

		select {
		case transport.Stream() <- entry:
		default:
			fmt.Println(errors.Errorf("%s has exeeded its capacity, dropping backpreassure", transport.Name()))
		}
		return true
	})
}

func (w *Writer) runPublisher() {
	w.wg.Add(1)
	w.eg.Go(func() error {
		defer w.closeTransportStreams()
		w.wg.Done()
		for entry := range w.queue {
			select {
			case <-w.ctx.Done():
				return nil
			default:
				w.publishToAllTransportStreams(entry)
			}
		}
		return nil
	})
}

// RegisterTransport accepts a writeFunc and kicks off a go routine which ensures delivery of log messages to the new writeFunc
// WriteFunc is used in a goroutine and is passed each log entry as they arrive.
// The behavior of the Writer is to attempt to deliver messages to each WriteFunc as fast as possible.
// If a WriteFunc cannot keep up and backpressure on the queue reaches 1000 messages will be dropped from this
// WriteFunc stream until it can handle additional capacity
// The WriteFunc should only return an error if it should no longer receive log entries
func (w *Writer) RegisterTransport(transport Transport, filters ...func(entry Entry) bool) error {
	if _, loaded := w.transports.LoadOrStore(transport.Name(), transport); loaded {
		return errors.Errorf("transport with name %s already exists", transport.Name())
	}

	w.wg.Add(1)
	w.eg.Go(func() error {
		defer func() {
			w.transports.Delete(transport.Name())
			if err := transport.Cleanup(); err != nil {
				fmt.Println(errors.Wrapf(err, "failed to cleanup logz transport %s", transport.Name()))
			}
		}()

		w.wg.Done()
		for threadSaveEntry := range transport.Stream() {

			entry := Entry{
				Fields:    make(Fields),
				Timestamp: threadSaveEntry.Timestamp,
				Level:     threadSaveEntry.Level,
				Caller:    threadSaveEntry.Caller,
				Message:   threadSaveEntry.Message,
				Error:     threadSaveEntry.Error,
			}

			for _, v := range threadSaveEntry.Fields {
				entry.Fields[v[0]] = v[1]
			}

			select {
			case <-w.ctx.Done():
				return nil
			default:
				for _, filter := range filters {
					if !filter(entry) {
						continue
					}
				}
				if err := transport.Write(entry); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return nil
}

// Wait waits for all transports in the error group to complete before returning
func (w *Writer) Wait() error {
	return w.eg.Wait()
}

// New builds a new *Writer
func New(ctx context.Context, opts ...Option) *Writer {
	w := &Writer{
		level:      InfoLevel,
		fields:     make([][]string, 0),
		queue:      make(chan ThreadSaveEntry, 1000),
		once:       new(sync.Once),
		wg:         new(sync.WaitGroup),
		transports: new(sync.Map),
	}

	w.eg, w.ctx = errgroup.WithContext(ctx)

	for _, opt := range opts {
		if err := opt(w); err != nil {
			w.error = multierror.Append(w.error, err)
		}
	}

	if !w.disableAutoConsume {
		w.start()
		w.wg.Wait()
	}

	return w
}
