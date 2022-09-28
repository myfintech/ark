package commands

import (
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
)

var (
	// GraphRunnerExecute the topics.GraphRunnerCommands execute command
	GraphRunnerExecute     = topics.GraphRunnerCommands.With("execute")
	GraphRunnerExecuteType = cqrs.WithType(GraphRunnerExecute)

	GraphRunnerCancel     = topics.GraphRunnerCommands.With("cancel")
	GraphRunnerCancelType = cqrs.WithType(GraphRunnerCancel)
)
