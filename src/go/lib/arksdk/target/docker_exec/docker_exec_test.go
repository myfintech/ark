package docker_exec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

var execTargetHCL = `
package "test" {
	description = ""
}
target "docker_exec" "test" {
	image = "ubuntu:latest"
	command = ["bash", "-c", "echo $PWD && cat /proc/version && ls -lha"]
}
target "docker_exec" "fail_test1" {
	image = "ubuntu:latest"
	command = ["bash", "-c", "ls -lah | grep -o fail"]
}
target "docker_exec" "fail_test2" {
	image = "ubuntu:latest"
	command = ["bash", "-c", "ls -lah"]
	ports = ["8080:8080:8080"]
}
target "docker_exec" "fail_test3" {
	image = "ubuntu:latest"
	command = ["bash", "-c", "ls -lah"]
	ports = ["8080:-8080"]
}
target "docker_exec" "fail_test4" {
	image = "ubuntu:latest"
	command = ["bash", "-c", "ls -lah"]
	ports = ["things:8080"]
}
`

func TestTargetExec_Build(t *testing.T) {
	cwd, _ := os.Getwd()
	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"docker_exec": Target{},
	}
	require.NoError(t, workspace.DetermineRootFromCWD())

	exampleBuildFile := filepath.Join(cwd, "BUILD.hcl")
	exampleHCLFile, diag := hclutils.FileFromString(execTargetHCL)
	if diag != nil && diag.HasErrors() {
		require.NoError(t, diag)
	}

	err := workspace.LoadTargets([]base.BuildFile{
		{HCL: exampleHCLFile, Path: exampleBuildFile},
	})
	require.NoError(t, err, "must load target hcl files into workspace")

	t.Run("Command Success", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.docker_exec.test"))
	})

	t.Run("Command Failure", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.docker_exec.fail_test1"))
	})
	t.Run("Bad Port Mapping", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.docker_exec.fail_test2"))
	})
	t.Run("Port Not In Valid Port Range", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.docker_exec.fail_test3"))
	})
	t.Run("Port Not A Number", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.docker_exec.fail_test4"))
	})
}

func walkByTarget(t *testing.T, workspace *base.Workspace, address string) error {
	intendedTarget, err := workspace.TargetLUT.LookupByAddress(address)
	if err != nil {
		return err
	}

	return workspace.GraphWalk(intendedTarget.Address(), func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)
		if preBuildErr := buildable.PreBuild(); preBuildErr != nil {
			return preBuildErr
		}
		if buildErr := buildable.Build(); buildErr != nil {
			return buildErr
		}

		execTarget := buildable.(Target)
		_, cacheable := buildable.(base.Cacheable)
		require.Equal(t, false, cacheable)
		require.NotEmpty(t, execTarget.ComputedAttrs().Command)
		require.NotEmpty(t, execTarget.Command, "GetStateAttrs.Command should not be empty.")
		return nil
	})
}
