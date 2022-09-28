package transports

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/stretchr/testify/require"
)

func TestDefaultIOWriter(t *testing.T) {
	ctx := context.Background()
	logger := logz.New(ctx, logz.WithMux(DefaultIOWriter))
	require.NoError(t, logger.InitError())

	for i := 0; i < 100; i++ {
		logger.Info("testing ", i)
	}
	logger.Close()
	require.NoError(t, logger.Wait())
}

func TestIOWriterWithFile(t *testing.T) {
	ctx := context.Background()
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")
	logFilePath := filepath.Join(testdata, "example.log")
	defer func() {
		_ = os.RemoveAll(logFilePath)
	}()

	logger := logz.New(ctx, logz.WithMux(DefaultIOWriter, DefaultFileWriter(logFilePath)))
	require.NoError(t, logger.InitError())

	for i := 0; i < 100; i++ {
		logger.Info("testing ", i)
	}

	logger.Close()
	require.NoError(t, logger.Wait())

	logFile, err := os.Open(logFilePath)
	require.NoError(t, err)

	reader := bufio.NewReader(logFile)
	defer func() {
		_ = logFile.Close()
	}()
	for i := 0; i < 100; i++ {
		line, _, lErr := reader.ReadLine()
		require.NoError(t, lErr)
		require.Contains(t, string(line), fmt.Sprint("testing ", i))
	}
}
