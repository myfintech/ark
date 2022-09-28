package http_handlers

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/myfintech/ark/src/go/lib/utils"
	"golang.org/x/sync/errgroup"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/logz"
)

type arkLogTesting struct {
	logPath     string
	logContents []byte
	addr        string
	server      *fiber.App
	waitGroup   *sync.WaitGroup
}

func (a *arkLogTesting) addLogPath(logPath string) {
	a.logPath = logPath
}

func (a *arkLogTesting) addLogContents(logContents []byte) {
	a.logContents = logContents
}

func (a *arkLogTesting) addAddress(addr string) {
	a.addr = addr
}

func (a *arkLogTesting) addServer(server *fiber.App) {
	a.server = server
}

var testSuite = &arkLogTesting{}

func init() {
	logger := logz.NoOpLogger{}
	testSuite.addLogContents([]byte(`i am a logfile
dont really care whats in here
as long as the contents match
once pulled from the server

this is a haiku
the lines need set syllables
refrigerator`))

	logPath, err := logz.SuggestedFilePath("ark/graph/someKey", "run.log")
	testSuite.addLogPath(logPath)
	if err != nil {
		panic(err)
	}

	logPathWithoutFile := filepath.Dir(testSuite.logPath)
	err = os.MkdirAll(logPathWithoutFile, 0770)
	if err != nil {
		panic(err)
	}
	arbitraryLogFile, err := os.Create(logPath)

	_, err = arbitraryLogFile.Write(testSuite.logContents)
	if err != nil {
		panic(err)
	}

	// Leave empty so that the path is forcibly pulled from the parameter
	logHandlerFunc := NewLogsHandler("", logger)
	testSuite.addServer(fiber.New(fiber.Config{
		DisableStartupMessage: true,
	}))

	testSuite.server.Get("/server/logs/:log_key", logHandlerFunc)

	ctx := context.Background()
	eg, _ := errgroup.WithContext(ctx)
	testSuite.waitGroup = new(sync.WaitGroup)
	testSuite.waitGroup.Add(1) // remember to add more if there are more test funcs that need to run

	port, _ := utils.GetFreePort()
	testSuite.addAddress(fmt.Sprintf("127.0.0.1:%s", port))

	waitForStart := make(chan struct{})
	eg.Go(func() error {
		close(waitForStart)
		testSuite.waitGroup.Done()
		logger.Infof("listening on http://%s", testSuite.addr)
		return testSuite.server.Listen(testSuite.addr)
	})
	<-waitForStart
	testSuite.waitGroup.Wait()
}

func TestNewLogsHandler(t *testing.T) {
	defer func() {
		err := os.Remove(testSuite.logPath)
		require.NoError(t, err)
	}()

	client := new(http.Client)
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/server/logs/someKey", testSuite.addr), nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, 200)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	respBytes, err := ioutil.ReadAll(io.LimitReader(resp.Body, int64(len(testSuite.logContents))))
	require.NoError(t, err)
	require.NotEmpty(t, respBytes)

	respContents := string(respBytes)
	require.Equal(t, string(testSuite.logContents), respContents)
}
