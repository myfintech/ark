package messages

import (
	"github.com/myfintech/ark/src/go/lib/fs"
)

// GraphRunnerExecuteCommand is a struct that represent the payload for the command handler request
type GraphRunnerExecuteCommand struct {
	TargetKeys     []string `json:"targetKeys"`
	K8sContext     string   `json:"k8sContext"`
	K8sNamespace   string   `json:"k8sNamespace"`
	PushAfterBuild bool     `json:"pushAfterBuild"`
	ForceBuild     bool     `json:"forceBuild"`
	SkipFilters    []string `json:"skipFilters"`
	MaxConcurrency int      `json:"maxConcurrency"`
}

// GraphRunnerExecuteCommandResponse is a struct that represent the payload for the command handler response
type GraphRunnerExecuteCommandResponse struct {
	SubscriptionId string `json:"subscriptionId"`
}

// FileSystemObserverFileChanged a struct
type FileSystemObserverFileChanged struct {
	Files []*fs.File
}
type K8sEchoResourceChangedAction int64

const (
	K8sEchoResourceChangedActionAdded
	K8sEchoResourceChangedActionUpdated
	K8sEchoResourceChangedActionDeleted
)

type K8sEchoResourceChanged struct {
	Name        string                       `json:"name"`
	Action      K8sEchoResourceChangedAction `json:"action"`
	Namespace   string                       `json:"namespace"`
	Labels      map[string]string            `json:"labels"`
	Annotations map[string]string            `json:"annotations"`
	Raw         string                       `json:"raw"`
}
