package health_check

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"go.temporal.io/sdk/activity"
)

//HeartbeatActivity attempts to connect to a given endpoint over TCP and HTTP and return the results of the connection
func HeartbeatActivity(ctx context.Context, opts HeartbeatOpts) (HeartbeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Heartbeat activity started")

	opts.socket = fmt.Sprintf("%s:%d", opts.Address, opts.Port)
	result := HeartbeatResult{
		TCPSuccess:   false,
		ResponseBody: "",
	}

	_, err := net.Dial("tcp", opts.socket)
	if err != nil {
		return result, err
	}
	result.TCPSuccess = true

	url := fmt.Sprintf("http://%s/%s", opts.socket, opts.HealthCheckPath)
	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	result.ResponseCode = resp.StatusCode

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	err = resp.Body.Close()
	if err != nil {
		return result, err
	}
	trimmedBody := strings.Trim(string(body), "\n")
	result.ResponseBody = trimmedBody

	return result, nil
}

//AlertActivity reaches out asynchronously to an alerting service after the heartbeat is unable to connect for a given time and number of attempts
func AlertActivity(ctx context.Context, _ HeartbeatOpts, _ error) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Alert activity started")

	fmt.Println("Placeholder function - send an alert to some endpoint or end user, tbd")

	logger.Info("Alert activity finished")
	return nil
}
