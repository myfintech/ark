package http_server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"

	"github.com/pkg/errors"
	"gopkg.in/h2non/gentleman.v2"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/dag"
)

// ClientGentleman defines the gentleman.ClientGentleman to be passed around
type ClientGentleman struct {
	client *gentleman.Client
}

// NewClient creates a new pointer to a gentleman http client
func NewClient(baseURL string) Client {
	return &ClientGentleman{
		client: gentleman.New().BaseURL(baseURL),
	}
}

// AddTarget adds a target to the database using the http client
func (c ClientGentleman) AddTarget(target ark.RawTarget) (ark.RawArtifact, error) {
	artifact := ark.RawArtifact{}
	res, err := c.client.Request().
		Path("/targets").
		Method("POST").
		JSON(target).
		Send()

	defer func() {
		_ = res.Close()
	}()

	if err != nil {
		return artifact, err
	}

	if !res.Ok {
		errRes := new(api_errors.APIError)
		if err = res.JSON(errRes); err != nil {
			return artifact, err
		}

		if errRes.Code == "" {
			return artifact, errors.Errorf("%d %s %s request failed",
				res.StatusCode,
				res.RawRequest.Method,
				res.RawRequest.URL,
			)
		}

		return artifact, errRes
	}

	return artifact, res.JSON(&artifact)
}

// GetTargets gets all of the targets from the database and returns them as a slice of ark.Targets
func (c ClientGentleman) GetTargets() ([]ark.RawTarget, error) {
	var targets []ark.RawTarget

	res, err := c.client.Request().
		Path("/targets").
		Method("GET").
		Send()

	defer func() {
		_ = res.Close()
	}()

	if err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, errors.Errorf("Request error: %v", res.StatusCode)
	}

	if err = res.JSON(&targets); err != nil {
		return nil, err
	}

	return targets, nil
}

// ConnectTargets connects the nodes in the graph
func (c ClientGentleman) ConnectTargets(edge ark.GraphEdge) (ark.GraphEdge, error) {
	// FIXME(chris): this isn't actually making an HTTP call
	return edge, nil
}

// GetGraph retrieves the graph from the database
func (c ClientGentleman) GetGraph() (*dag.AcyclicGraph, error) {
	graph := new(dag.AcyclicGraph)

	res, err := c.client.Request().
		Path("/graph").
		Method("GET").
		Send()

	defer func() {
		_ = res.Close()
	}()

	if err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, errors.Errorf("Request error: %v", res.StatusCode)
	}

	if err = res.JSON(graph); err != nil {
		return graph, err
	}

	return graph, nil
}

// GetGraphEdges retrieves all of the graph edges from the database
func (c ClientGentleman) GetGraphEdges() ([]ark.GraphEdge, error) {
	var graphEdges []ark.GraphEdge

	res, err := c.client.Request().
		Path("/graph/edges").
		Method("GET").
		Send()

	defer func() {
		_ = res.Close()
	}()

	if err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, errors.Errorf("Request error: %v", res.StatusCode)
	}

	return graphEdges, res.JSON(&graphEdges)
}

// Run execute targets
func (c ClientGentleman) Run(request messages.GraphRunnerExecuteCommand) (cmdRes messages.GraphRunnerExecuteCommandResponse, err error) {

	res, err := c.client.Request().
		Path("/run/").
		Method(http.MethodPost).
		JSON(request).
		Send()

	defer func() {
		_ = res.Close()
	}()

	if err != nil {
		return
	}

	if !res.Ok {
		errRes := new(api_errors.APIError)
		if err = res.JSON(errRes); err != nil {
			return cmdRes, err
		}

		if errRes.Code == "" {
			return cmdRes, errors.Errorf("Request error: %v", res.StatusCode)
		}

		return cmdRes, errRes
	}

	err = res.JSON(&cmdRes)

	return
}

// GetServerLogs streams all server logs
func (c ClientGentleman) GetServerLogs() (io.Reader, error) {
	res, err := c.client.Request().
		Path("/server/logs").
		Method(http.MethodGet).
		Send()
	if err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, errors.Errorf("Request error: %v", res.StatusCode)
	}

	return res, nil
}

// GetLogsByKey streams logs for a specific target's key
func (c ClientGentleman) GetLogsByKey(logKey string) (io.Reader, error) {
	res, err := c.client.Request().
		Path(fmt.Sprintf("/server/logs/%s", logKey)).
		Method(http.MethodGet).
		Send()
	if err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, errors.Errorf("Request error: %v", res.StatusCode)
	}

	return res, nil
}
