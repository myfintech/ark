package stdlib

import (
	"github.com/dop251/goja"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/myfintech/ark/src/go/lib/watchman"
)

type Client interface {
	AddTarget(target ark.RawTarget) (ark.RawArtifact, error)
	GetTargets() ([]ark.RawTarget, error)
	ConnectTargets(edge ark.GraphEdge) (ark.GraphEdge, error)
	GetGraph() (*dag.AcyclicGraph, error)
	GetGraphEdges() ([]ark.GraphEdge, error)
}

type Options struct {
	FSRealm        string
	Client         Client
	Runtime        *goja.Runtime
	GitIgnore      gitignore.Matcher
	WatchmanClient *watchman.Client
}
