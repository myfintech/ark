package health_check

import (
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/testsuite"
)

func TestHealthCheckWorkflow(t *testing.T) {
	successSuite := &testsuite.WorkflowTestSuite{}
	env := successSuite.NewTestWorkflowEnvironment()

	successfulCheck := HeartbeatResult{
		TCPSuccess:   true,
		ResponseBody: "",
		ResponseCode: 200,
	}
	hbOpts := HeartbeatOpts{
		Timeout:            0,
		HeartbeatInterval:  5 * time.Second,
		NotifyInterval:     5 * time.Second,
		UnhealthyThreshold: 2,
		Port:               30005,
		Address:            "127.0.0.1",
	}

	env.OnActivity(HeartbeatActivity, mock.Anything, hbOpts).Return(successfulCheck, nil)
	env.ExecuteWorkflow(HeartbeatWorkflow, hbOpts)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result HeartbeatResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, successfulCheck, result)

	failSuite := &testsuite.WorkflowTestSuite{}
	env = failSuite.NewTestWorkflowEnvironment()
	failCheck := HeartbeatResult{
		TCPSuccess:   false,
		ResponseBody: "",
		ResponseCode: 0,
	}
	activityErr := errors.New("Endpoint timeout")
	env.OnActivity(HeartbeatActivity, mock.Anything, hbOpts).Return(failCheck, activityErr)
	env.OnActivity(AlertActivity, mock.Anything, hbOpts).Return(nil)
	env.ExecuteWorkflow(HeartbeatWorkflow, hbOpts)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
