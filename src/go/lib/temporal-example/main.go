package main

import (
	"context"
	"fmt"
	"time"

	"github.com/myfintech/ark/src/go/lib/temporal-example/health_check"
	"go.temporal.io/sdk/client"
)

func main() {
	// Make sure docker compose is running the temporal server before running
	temporal, err := client.NewClient(client.Options{})
	if err != nil {
		panic(err)
	}
	defer temporal.Close()
	workflowOpts := client.StartWorkflowOptions{
		ID:        "health-check-workflow",
		TaskQueue: health_check.TaskQueue,
	}
	hcOpts := health_check.HeartbeatOpts{
		Timeout:            30 * time.Second,
		HeartbeatInterval:  5 * time.Minute,
		NotifyInterval:     30 * time.Minute,
		UnhealthyThreshold: 3,
		Port:               8888,
		Address:            "127.0.0.1",
	}
	hcWorkflow, err := temporal.ExecuteWorkflow(context.Background(), workflowOpts, health_check.HeartbeatWorkflow, hcOpts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nWorkflowID: %s RunID: %s\n", hcWorkflow.GetID(), hcWorkflow.GetRunID())
}
