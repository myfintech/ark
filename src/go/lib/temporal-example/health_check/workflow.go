package health_check

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

//TaskQueue is the shared name that should be used by workers, workflows and activities
const TaskQueue = "HEALTH_CHECK_TASK_QUEUE"

//HeartbeatOpts configures the intervals and other options of a given health check
type HeartbeatOpts struct {
	Timeout            time.Duration
	HeartbeatInterval  time.Duration
	NotifyInterval     time.Duration
	UnhealthyThreshold int32
	Port               uint
	Address            string
	HealthCheckPath    string
	socket             string
}

//HeartbeatFailureQuery tracks the number of consecutive failures
type HeartbeatFailureQuery struct {
	Count int
}

//HeartbeatResult wraps the information returned by an endpoint's health check
type HeartbeatResult struct {
	TCPSuccess   bool
	ResponseBody string
	ResponseCode int
}

//HeartbeatWorkflow pushes a heartbeat activity to a task queue that can be picked up by a ready worker
func HeartbeatWorkflow(ctx workflow.Context, hbOpts HeartbeatOpts) (HeartbeatResult, error) {
	activityOpts := workflow.ActivityOptions{
		TaskQueue:           TaskQueue,
		StartToCloseTimeout: 20 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    hbOpts.HeartbeatInterval,
			BackoffCoefficient: 1,
			MaximumAttempts:    hbOpts.UnhealthyThreshold,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	logger := workflow.GetLogger(ctx)
	logger.Info("Heartbeat Workflow started")

	var result HeartbeatResult
	err := workflow.ExecuteActivity(ctx, HeartbeatActivity, hbOpts).Get(ctx, &result)
	if err != nil {
		emptyResult := HeartbeatResult{}
		logger.Info("Heartbeat activity exceeded failure threshold, notifying emergency contact")
		workflow.ExecuteActivity(ctx, AlertActivity, hbOpts, err)
		return emptyResult, nil
	}

	logger.Info("Heartbeat Workflow completed")
	return result, nil
}
