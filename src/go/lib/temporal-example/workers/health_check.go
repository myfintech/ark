package main

import (
	"github.com/myfintech/ark/src/go/lib/temporal-example/health_check"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	temporal, err := client.NewClient(client.Options{})
	if err != nil {
		panic(err)
	}
	defer temporal.Close()

	w := worker.New(temporal, health_check.TaskQueue, worker.Options{})
	w.RegisterWorkflow(health_check.HeartbeatWorkflow)
	w.RegisterActivity(health_check.HeartbeatActivity)
	w.RegisterActivity(health_check.AlertActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		panic(err)
	}
}
