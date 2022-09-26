package derivation

import (
	"github.com/mitchellh/mapstructure"
	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/targets/deploy"
	"github.com/myfintech/ark/src/go/lib/ark/targets/docker_image"
	"github.com/myfintech/ark/src/go/lib/ark/targets/group"
	"github.com/myfintech/ark/src/go/lib/ark/targets/kube_exec"
	"github.com/myfintech/ark/src/go/lib/ark/targets/local_file"
	"github.com/myfintech/ark/src/go/lib/ark/targets/nix"
	"github.com/myfintech/ark/src/go/lib/ark/targets/probe"
	"github.com/myfintech/ark/src/go/lib/ark/targets/sync_kv"
	"github.com/myfintech/ark/src/go/lib/ark/targets/test"
	"github.com/pkg/errors"
)

// TargetFromRawTarget maps an ark.RawTarget to an ark.Target by type
func TargetFromRawTarget(rawTarget ark.RawTarget) (target ark.Target, err error) {
	switch rawTarget.Type {
	case docker_image.Type:
		target = &docker_image.Target{RawTarget: rawTarget}
	case deploy.Type:
		target = &deploy.Target{RawTarget: rawTarget}
	case sync_kv.Type:
		target = &sync_kv.Target{RawTarget: rawTarget}
	case kube_exec.Type:
		target = &kube_exec.Target{RawTarget: rawTarget}
	case group.Type:
		target = &group.Target{RawTarget: rawTarget}
	case probe.Type:
		target = &probe.Target{RawTarget: rawTarget}
	case test.Type:
		target = &test.Target{RawTarget: rawTarget}
	case local_file.Type:
		target = &local_file.Target{RawTarget: rawTarget}
	case nix.Type:
		target = &nix.Target{RawTarget: rawTarget}
	default:
		return nil, errors.Errorf("invalid target type %s", rawTarget.Type)
	}

	if err = mapstructure.Decode(rawTarget.Attributes, target); err != nil {
		return
	}
	return
}

// TargetAndArtifactFromRawTarget derives an ark.Artifact from an ark.RawTarget by type
func TargetAndArtifactFromRawTarget(rawTarget ark.RawTarget) (target ark.Target, artifact ark.Artifact, err error) {
	checksum, err := rawTarget.Checksum()
	if err != nil {
		return
	}

	target, err = TargetFromRawTarget(rawTarget)
	if err != nil {
		return
	}

	artifact, err = target.Produce(checksum)
	if err != nil {
		return
	}

	return
}

// RawArtifactFromRawTarget derives an ark.RawArtifact from an ark.RawTarget by type
func RawArtifactFromRawTarget(rawTarget ark.RawTarget) (rawArtifact ark.RawArtifact, err error) {
	_, artifact, err := TargetAndArtifactFromRawTarget(rawTarget)
	if err != nil {
		return
	}

	return RawArtifactFromArtifact(artifact)
}

// RawArtifactFromArtifact maps an ark.Artifact to an ark.RawArtifact
func RawArtifactFromArtifact(artifact ark.Artifact) (rawArtifact ark.RawArtifact, err error) {
	err = mapstructure.Decode(artifact, &rawArtifact)
	return
}

// ActionFromTargetAndArtifact maps an ark.Target to an ark.Action by type
func ActionFromTargetAndArtifact(target ark.Target, artifact ark.Artifact) (action ark.Action, err error) {
	defer func() {
		if v := recover(); v != nil {
			err = v.(error)
		}
	}()
	switch t := target.(type) {
	case *docker_image.Target:
		action = &docker_image.Action{Target: t, Artifact: artifact.(*docker_image.Artifact)}
	case *deploy.Target:
		action = &deploy.Action{Target: t, Artifact: artifact.(*deploy.Artifact)}
	case *sync_kv.Target:
		action = &sync_kv.Action{Target: t, Artifact: artifact.(*sync_kv.Artifact)}
	case *kube_exec.Target:
		action = &kube_exec.Action{Target: t, Artifact: artifact.(*kube_exec.Artifact)}
	case *group.Target:
		action = &group.Action{Target: t, Artifact: artifact.(*group.Artifact)}
	case *probe.Target:
		action = &probe.Action{Target: t, Artifact: artifact.(*probe.Artifact)}
	case *test.Target:
		action = &test.Action{Target: t, Artifact: artifact.(*test.Artifact)}
	case *local_file.Target:
		action = &local_file.Action{Target: t, Artifact: artifact.(*local_file.Artifact)}
	case *nix.Target:
		action = nix.Action{Target: t, Artifact: artifact.(*nix.Artifact)}
	}
	return
}
