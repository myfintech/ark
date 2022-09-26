package health_check

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

//tcpCheck.Get method is returning an internal value, something weird is happening. investigate later
func TestHealthCheckActivity(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(HeartbeatActivity)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "OK")
		require.NoError(t, err)
	}))
	defer testServer.Close()
	substrings := strings.Split(testServer.URL, ":")
	ip := strings.TrimLeft(substrings[1], "//")
	port, err := strconv.Atoi(substrings[2])
	require.NoError(t, err)

	opts := HeartbeatOpts{
		Timeout:            0,
		HeartbeatInterval:  0,
		NotifyInterval:     0,
		UnhealthyThreshold: 0,
		Port:               uint(port),
		Address:            ip,
	}
	tcpCheck, err := env.ExecuteActivity(HeartbeatActivity, opts)
	require.NoError(t, err)

	var tcpResult HeartbeatResult
	expectedHealthyOutput := HeartbeatResult{
		TCPSuccess: true,
	}
	require.NoError(t, tcpCheck.Get(&tcpResult))
	require.Equal(t, expectedHealthyOutput.TCPSuccess, tcpResult.TCPSuccess)

	httpCheck, err := env.ExecuteActivity(HeartbeatActivity, opts)
	require.NoError(t, err)

	var httpResult HeartbeatResult
	expectedHTTPOutput := HeartbeatResult{
		TCPSuccess:   true,
		ResponseBody: "OK",
		ResponseCode: 200,
	}
	require.NoError(t, httpCheck.Get(&httpResult))
	require.Equal(t, expectedHTTPOutput, httpResult)

}
