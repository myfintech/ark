package cmd

import (
	"github.com/myfintech/ark/src/go/lib/arksdk/target/deploy"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/group"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/kv_sync"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/nix"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/probe"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/test"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/build"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/docker_exec"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/docker_image"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/http_archive"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/jsonnet"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/jsonnet_file"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/kube_exec"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/local_exec"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/local_file"
)

func decodeWorkspaceOnlyPreRunE(cmd *cobra.Command, args []string) error {
	workspace.ExtractCliOptions(cmd, args)

	if err := workspace.DetermineRootFromCWD(); err != nil {
		return errors.Wrap(err, "there was an error getting the workspace root directory")
	}

	if err := workspace.DecodeFile(nil); err != nil {
		return errors.Wrap(err, "there was an error decoding the WORKSPACE.hcl file")
	}

	if err := workspace.InitKubeClient(viper.GetString("namespace")); err != nil {
		return errors.Wrap(err, "failed to init kubernetes client")
	}

	if err := workspace.InitVaultClient(); err != nil {
		return errors.Wrap(err, "failed to init Vault client")
	}

	if err := workspace.InitDockerClient(); err != nil {
		return errors.Wrap(err, "failed to init docker client")
	}
	return nil
}

func decodeWorkspacePreRunE(cmd *cobra.Command, args []string) error {
	workspace.ExtractCliOptions(cmd, args)
	workspace.SetEnvironmentConstraint(viper.GetString("environment"))

	workspace.RegisteredTargets = base.Targets{
		"docker_image": docker_image.Target{},
		"build":        build.Target{},
		"docker_exec":  docker_exec.Target{},
		"exec":         local_exec.Target{},
		"jsonnet":      jsonnet.Target{},
		"jsonnet_file": jsonnet_file.Target{},
		"http_archive": http_archive.Target{},
		"local_file":   local_file.Target{},
		"probe":        probe.Target{},
		"kube_exec":    kube_exec.Target{},
		"deploy":       deploy.Target{},
		"group":        group.Target{},
		"kv_sync":      kv_sync.Target{},
		"test":         test.Target{},
		"nix":          nix.Target{},
	}

	if err := workspace.DetermineRootFromCWD(); err != nil {
		return errors.Wrap(err, "there was an error getting the workspace root directory")
	}

	if err := workspace.DecodeFile(nil); err != nil {
		return errors.Wrap(err, "there was an error decoding the WORKSPACE.hcl file")
	}

	if err := workspace.InitKubeClient(viper.GetString("namespace")); err != nil {
		return errors.Wrap(err, "failed to init kubernetes client")
	}

	if err := workspace.InitVaultClient(); err != nil {
		return errors.Wrap(err, "failed to init Vault client")
	}

	if err := workspace.InitDockerClient(); err != nil {
		return errors.Wrap(err, "failed to init docker client")
	}

	buildFiles, err := workspace.DecodeBuildFiles()
	if err != nil {
		return errors.Wrap(err, "there was an error decoding the workspace build files")
	}

	if err = workspace.LoadTargets(buildFiles); err != nil {
		return err
	}

	// go func() {
	// 	_ = portbinder.New(workspace.Context, workspace.K8s, workspace.PortBinderCommands, workspace.ReadyPortCommands)
	// }()

	return nil
}
