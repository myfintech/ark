package host_server

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/log"
)

// func loadWorkspaceFromRoot(dir string) error {
//	workspace.RegisteredTargets = base.Targets{
//		"docker_image": docker_image.Target{},
//		"build":        build.Target{},
//		"docker_exec":  docker_exec.Target{},
//		"exec":         local_exec.Target{},
//		"jsonnet":      jsonnet.Target{},
//		"jsonnet_file": jsonnet_file.Target{},
//		"http_archive": http_archive.Target{},
//		"local_file":   local_file.Target{},
//		"kube_exec":    kube_exec.Target{},
//	}
//
//	if err := workspace.DetermineRoot(dir); err != nil {
//		return errors.Wrap(err, "there was an error getting the workspace")
//	}
//
//	if err := workspace.DecodeFile(nil); err != nil {
//		return errors.Wrap(err, "there was an error decoding the WORKSPACE.hcl file")
//	}
//
//	if err := workspace.InitKubeClient(); err != nil {
//		return errors.Wrap(err, "failed to init kubernetes client")
//	}
//
//	if err := workspace.InitVaultClient(); err != nil {
//		return errors.Wrap(err, "failed to init Vault client")
//	}
//
//	if err := workspace.InitDockerClient(); err != nil {
//		return errors.Wrap(err, "failed to init docker client")
//	}
//
//	buildFiles, err := workspace.DecodeBuildFiles()
//	if err != nil {
//		return errors.Wrap(err, "there was an error decoding the workspace build files")
//	}
//
//	if err = workspace.LoadTargets(buildFiles); err != nil {
//		return err
//	}
//
//	return nil
// }

// ArkStore represents a set of fields that are global to all workspaces
type ArkStore struct{}

// GetArtifactsDir returns the artifacts directory
func (s ArkStore) GetArtifactsDir() string {
	return filepath.Join(s.GetRootDir(), "artifacts")
}

// GetRootDir returns the /ark directory
func (s ArkStore) GetRootDir() string {
	return "/ark"
}

// Clean removes generated files from the /ark/artifacts directory
// this is used to reset the local state and artifacts cache
func (s ArkStore) Clean() error {
	return os.RemoveAll(s.GetArtifactsDir())
}

// hostServer implements the HostServer interface
type hostServer struct {
	workspaces sync.Map
	store      ArkStore
}

// New returns a new instance of the host server

// GetWorkspace returns a pointer to a workspace object based on the provided workspace root
func (h *hostServer) GetWorkspace(root string) (*base.Workspace, error) {
	workspace, ok := h.workspaces.Load(root)
	if !ok {
		return &base.Workspace{}, errors.New("the provided workspace root is not valid")
	}

	return workspace.(*base.Workspace), nil
}

// CleanHostArtifacts removes the locally stored artifacts
func (h *hostServer) CleanHostArtifacts(_ context.Context, _ *CleanHostArtifactsRequest) (*CleanHostArtifactsResponse, error) {
	return nil, h.store.Clean()
}

// PullArtifacts pulls any remotely cached artifacts
func (h *hostServer) PullArtifacts(_ *PullArtifactsRequest, _ Host_PullArtifactsServer) error {
	return status.Errorf(codes.Unimplemented, "method PullArtifacts not implemented")
}

// PushArtifacts pushes any locally built artifacts to remote storage
func (h *hostServer) PushArtifacts(_ *PushArtifactsRequest, _ Host_PushArtifactsServer) error {
	return status.Errorf(codes.Unimplemented, "method PushArtifacts not implemented")
}

// ListTargets outputs a list of available targets for a workspace
func (h *hostServer) ListTargets(_ context.Context, req *ListTargetsRequest) (*ListTargetsResponse, error) {
	var response = ListTargetsResponse{}

	workspace, err := h.GetWorkspace(req.Root)
	if err != nil {
		return &response, err
	}

	for _, address := range workspace.TargetLUT.SortedAddresses() {
		addressable := workspace.TargetLUT[address]
		buildableTarget, buildable := addressable.(base.Buildable)
		cacheableTarget, cacheable := addressable.(base.Cacheable)
		locallyCached := false
		remotelyCached := false

		if buildable {
			if preBuildErr := buildableTarget.PreBuild(); preBuildErr != nil {
				return &response, preBuildErr
			}
		}
		if cacheable {
			var cacheErr error
			locallyCached, cacheErr = cacheableTarget.CheckLocalBuildCache()
			if cacheErr != nil {
				return &response, cacheErr
			}

			if req.CheckRemoteCache {
				remotelyCached, cacheErr = cacheableTarget.CheckRemoteCache()
				if cacheErr != nil {
					return &response, cacheErr
				}
			}
		}
		response.Targets = append(response.Targets, &Target{
			Address:        address,
			ShortHash:      buildableTarget.ShortHash(),
			Hash:           buildableTarget.Hash(),
			LocallyCached:  locallyCached,
			RemotelyCached: remotelyCached,
			Description:    addressable.Describe(),
		})
	}

	return &response, nil
}

// RunTarget executes a target address
func (h *hostServer) RunTarget(req *RunTargetRequest, _ Host_RunTargetServer) error {
	workspace, err := h.GetWorkspace(req.Root)
	if err != nil {
		return err
	}

	rootTarget, err := workspace.TargetLUT.LookupByAddress(req.Address)
	if err != nil {
		return err
	}

	req.Push = req.Push && workspace.Config.Artifacts != nil
	req.Pull = req.Pull && workspace.Config.Artifacts != nil

	if err = workspace.GraphWalk(rootTarget.Address(), base.BuildWalker(req.Force, req.Pull, req.Push)); err != nil {
		return err
	}

	if req.Watch {
		observable := workspace.WatchForChanges(
			base.FilterChangeNotificationsByTarget(workspace.TargetGraph.Isolate(rootTarget), rootTarget),
			base.ReRunOnChange(func() error {
				return workspace.GraphWalk(rootTarget.Address(), base.BuildWalker(req.Force, req.Pull, req.Push))
			}),
		)

		for message := range observable.Observe() {
			if message.Error() {
				log.Error(message.E)
				if req.StopOnFirstError {
					return message.E
				}
			}
		}
	}
	return nil
}

// AddWorkspace adds a workspace root to the server
func (h *hostServer) AddWorkspace(_ context.Context, _ *AddWorkspaceRequest) (*AddWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddWorkspace not implemented")
}

// ValidateWorkspace determines whether a workspace implementation is valid
func (h *hostServer) ValidateWorkspace(_ context.Context, _ *ValidateWorkspaceRequest) (*ValidateWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateWorkspace not implemented")
}

// WatchWorkspace establishes a watch stream with the configured file watcher
func (h *hostServer) WatchWorkspace(_ *WatchWorkspaceRequest, _ Host_WatchWorkspaceServer) error {
	return status.Errorf(codes.Unimplemented, "method WatchWorkspace not implemented")
}

// Shutdown instructs the server to shut itself down
func (h *hostServer) Shutdown(_ context.Context, _ *ShutdownRequest) (*ShutdownResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shutdown not implemented")
}
