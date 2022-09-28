package logz

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	tp := &mockTransport{
		name:   "t1",
		stream: make(chan ThreadSaveEntry, 1000),
	}

	tp.On("Cleanup")

	log := New(
		context.Background(),
		WithLevel(TraceLevel),
		WithMux(func() (transport Transport, err error) {
			return tp, nil
		}),
	)
	require.NoError(t, log.InitError())

	var stdMsgTestTable = []struct {
		opts     []interface{}
		expected string
		format   string
	}{
		{
			format:   "%d %d%s",
			opts:     []interface{}{1, 2, " test"},
			expected: "1 2 test",
		},
		{
			format:   "%v",
			opts:     []interface{}{errors.New("test")},
			expected: "test",
		},
	}

	var stdMethodTestTable = map[string]struct {
		level  Level
		method interface{}
	}{
		"Error":     {ErrorLevel, log.Error},
		"Error#fmt": {ErrorLevel, log.Errorf},

		"Warn":     {WarnLevel, log.Warn},
		"Warn#fmt": {WarnLevel, log.Warnf},

		"Info":     {InfoLevel, log.Info},
		"Info#fmt": {InfoLevel, log.Infof},

		"Debug":     {DebugLevel, log.Debug},
		"Debug#fmt": {DebugLevel, log.Debugf},

		"Trace":     {TraceLevel, log.Trace},
		"Trace#fmt": {TraceLevel, log.Tracef},
	}

	t.Run("should be able to log in multiple go routines", func(t *testing.T) {
		for _, test := range stdMsgTestTable {
			for name, methodTest := range stdMethodTestTable {
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					tp.On("Write", test.expected, methodTest.level).Return(nil)
					switch m := methodTest.method.(type) {
					case func(...interface{}):
						m(test.opts...)
					case func(string, ...interface{}):
						m(test.format, test.opts...)
					default:
						t.Errorf("%v %T is not a valid log method type", m, m)
					}
				})
			}
		}
	})

	log.Close()
	err := log.Wait()
	require.NoError(t, err)
	tp.AssertExpectations(t)
}

func TestWriterWithFields(t *testing.T) {
	var entries = make(chan ThreadSaveEntry, 1024)
	var log FieldLogger = New(
		context.Background(),
		WithQueue(entries),
		WithLevel(TraceLevel),
		WithDisableAutoConsume(),
	)

	expectedFields := Fields{
		"key": "value",
	}

	log = log.Child(
		WithFields(expectedFields),
		WithDisableAutoConsume(),
	)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	t.Run("log entries should contain fields", func(t *testing.T) {
		log.Info("testing")
		entry := <-entries
		fields := make(Fields)
		for _, v := range entry.Fields {
			fields[v[0]] = v[1]
		}
		require.Equal(t, expectedFields, fields)
	})

	t.Run("trace level should include a runtime caller", func(t *testing.T) {
		log.Trace("testing")
		entry := <-entries
		// FIXME: this is not good. if we add more code we have to update the test line number
		require.Equal(t, filepath.Join(cwd, "writer_test.go:124"), entry.Caller)
	})

	t.Run("non trace level entries should not include a runtime caller", func(t *testing.T) {
		log.Info("testing")
		entry := <-entries
		require.Equal(t, entry.Caller, "")
	})
}
