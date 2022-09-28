package graph

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/myfintech/ark/src/go/lib/logz/transports"

	"github.com/myfintech/ark/src/go/lib/ark/targets/deploy"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	"github.com/myfintech/ark/src/go/lib/ark/targets/docker_image"

	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/storage/memory"
	"github.com/stretchr/testify/require"
)

var store ark.Store

func TestExecutionEngine(t *testing.T) {
	ctx := context.TODO()
	store = new(memory.Store)
	sharedClients, err := shared_clients.NewContainerWithDefaults()
	require.NoError(t, err)

	rootTarget := ark.RawTarget{
		Name:  "test",
		Type:  docker_image.Type,
		File:  "test",
		Realm: "test",
		Attributes: map[string]interface{}{
			"breakCache": time.Now(),
			"repo":       "gcr.io/[insert-google-project]/ark/tests/graph",
			"dockerfile": "FROM node:latest",
		},
		SourceFiles: nil,
		DependsOn:   nil,
	}

	_, err = store.AddTarget(rootTarget)

	// _, err := store.AddTarget(ark.RawTarget{
	// 	Name:        "test",
	// 	Type:        deploy.Type,
	// 	File:        "test",
	// 	Realm:       "test",
	// 	Attributes:  nil,
	// 	SourceFiles: nil,
	// 	DependsOn:   nil,
	// })

	require.NoError(t, err)

	logPath, err := logz.SuggestedFilePath("test", "thing")
	require.NoError(t, err)
	logger := logz.New(ctx, logz.WithMux(transports.DefaultFileWriter(logPath)))

	opts := ExecuteOptions{
		Ctx:            ctx,
		Store:          store,
		SharedClients:  sharedClients,
		RootTargetKey:  rootTarget.Key(),
		Broker:         new(cqrs.NoOpBroker),
		SubscriptionID: "test",
		Logger:         logger,
	}
	require.NoError(t, Execute(opts))
}

func TestValidationWalk(t *testing.T) {
	opts := ExecuteOptions{
		K8sNamespace: "",
	}
	walkerFunc := validationWalk(opts)
	require.Error(t, walkerFunc(&ark.RawTarget{Type: deploy.Type}))
}

type mockAction struct {
	Logger logz.FieldLogger
}

func (m *mockAction) UseLogger(logger logz.FieldLogger) {
	m.Logger = logger
}

type mockBadAction struct {
}

func Test_injectOrSkipLogger(t *testing.T) {
	ctx := context.TODO()
	logPath, err := logz.SuggestedFilePath("test", "thing")
	require.NoError(t, err)
	logger := logz.New(ctx, logz.WithMux(transports.DefaultFileWriter(logPath)))
	sharedClients, err := shared_clients.NewContainerWithDefaults()
	require.NoError(t, err)

	happyStruct := new(mockAction)

	input := injectOrSkipLoggerInput{
		action:         happyStruct,
		logger:         logger,
		subscriptionID: "some-random-guid",
		key:            "some-random-target-key",
		name:           "some-random-target-name",
		hashCode:       "some-random-artifact-hash-code",
		dockerClient:   sharedClients.Docker,
		k8sClient:      sharedClients.K8s,
	}
	injectOrSkipLogger(input)
	require.NotNil(t, happyStruct.Logger, fmt.Sprintf("Struct.logger value is: %s", happyStruct.Logger))

	require.FileExists(t, logPath)
	err = os.Remove(logPath)
	require.NoError(t, err)

	iSureHopeThisDoesntPanic := new(mockBadAction)
	input.action = iSureHopeThisDoesntPanic
	require.NotPanics(t, func() { injectOrSkipLogger(input) })
}
