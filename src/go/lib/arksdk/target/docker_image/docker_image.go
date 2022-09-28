package docker_image

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/container"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/exec"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

// Target defines the required and optional attributes for defining a Docker Image Build
type Target struct {
	*base.RawTarget `json:"-"`

	Repo          hcl.Expression `hcl:"repo,attr"`
	Dockerfile    hcl.Expression `hcl:"dockerfile,attr"` // Has to be relative to the repo/workspace root
	DockerContext hcl.Expression `hcl:"context,attr"`
	BuildArgs     hcl.Expression `hcl:"build_args,attr"`
	Tags          hcl.Expression `hcl:"tags,attr"`
	Output        hcl.Expression `hcl:"output,attr"`
	Progress      hcl.Expression `hcl:"progress,attr"`
}

// ComputedAttrs used to store the computed attributes of a docker_image target
type ComputedAttrs struct {
	Repo          string             `hcl:"repo,attr"`
	Dockerfile    string             `hcl:"dockerfile,attr"` // Has to be relative to the repo/workspace root
	DockerContext *string            `hcl:"context,attr"`
	BuildArgs     *map[string]string `hcl:"build_args,attr"`
	Tags          *[]string          `hcl:"tags,attr"`
	Output        *string            `hcl:"output,attr"`
	Progress      *string            `hcl:"progress,attr"`
}

// Attributes return combined rawTarget.Attributes with typedTarget.Attributes.
func (t Target) Attributes() map[string]cty.Value {
	computed := t.ComputedAttrs()
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{
		"repo": cty.StringVal(computed.Repo),
		"url":  cty.StringVal(fmt.Sprintf("%s:%s", computed.Repo, t.Hash())),
	})
}

// ComputedAttrs returns a pointer to computed attributes from the state store.
// If attributes are not in the state store it will create a new pointer and insert it into the state store.
func (t Target) ComputedAttrs() *ComputedAttrs {
	if attrs, ok := t.GetStateAttrs().(*ComputedAttrs); ok {
		return attrs
	}

	attrs := &ComputedAttrs{}

	t.SetStateAttrs(attrs)
	return attrs
}

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	return hclutils.DecodeExpressions(&t, t.ComputedAttrs(), base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
}

// Build constructs a Docker image from the information provided in the docker_image target
func (t Target) Build() error {
	attrs := t.ComputedAttrs()
	dockerArgs := []string{"build"}

	if attrs.DockerContext == nil {
		attrs.DockerContext = &t.Workspace.Dir
	}

	if !filepath.IsAbs(attrs.Dockerfile) {
		attrs.Dockerfile = filepath.Clean(filepath.Join(*attrs.DockerContext, attrs.Dockerfile))
	}

	imageTags := t.URLsFromTags()
	for _, tag := range imageTags {
		dockerArgs = append(dockerArgs, "-t", tag)
	}

	dockerArgs = append(dockerArgs, "-f", attrs.Dockerfile)

	if attrs.BuildArgs != nil {
		for k, v := range *attrs.BuildArgs {
			constructedArg := fmt.Sprintf("%s=%s", k, v)
			dockerArgs = append(dockerArgs, "--build-arg", constructedArg)
		}
	}

	if attrs.Output != nil {
		if err := os.MkdirAll(*attrs.Output, 0755); err != nil {
			return err
		}
		dockerArgs = append(dockerArgs, "--output", *attrs.Output)
	}

	if attrs.Progress != nil {
		dockerArgs = append(dockerArgs, "--progress", *attrs.Progress)
	}

	dockerArgs = append(dockerArgs, *attrs.DockerContext)

	cmd := exec.LocalExecutor(exec.LocalExecOptions{
		Command: append([]string{
			"docker",
		}, dockerArgs...),
		Dir:              "./",
		Stdin:            os.Stdin,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		InheritParentEnv: false,
	})

	log.Info(cmd.String())
	return cmd.Run()
}

// URL constructs an image URL from the repo and a given tag
func (t Target) URL(tag string) string {
	return fmt.Sprintf("%s:%s", t.ComputedAttrs().Repo, tag)
}

// URLsFromTags creates a slice of image URLs
func (t Target) URLsFromTags() []string {
	imageTags := []string{
		t.URL(t.Hash()),
		t.URL(t.ShortHash()),
		t.URL("latest"),
	}

	if t.ComputedAttrs().Tags != nil {
		for _, tag := range *t.ComputedAttrs().Tags {
			imageTags = append(imageTags, t.URL(tag))
		}
	}
	return imageTags
}

// CheckLocalBuildCache loads the build cache state and verifies that the hashes match
func (t Target) CheckLocalBuildCache() (bool, error) {
	ctx := context.Background()
	docker, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return false, errors.Wrap(err, "unable to create new docker client")
	}
	log.Debugf("checking for local Docker image: %s", t.URL(t.Hash()))
	exists, checkErr := docker.ImageExists(ctx, t.URL(t.Hash()))
	if checkErr != nil {
		return false, errors.Wrap(err, "CheckLocalBuildCache#ImageExists")
	}
	return exists, nil
}

// CheckRemoteCache determines whether a docker image for a given target exists in a remote repository
func (t Target) CheckRemoteCache() (bool, error) {
	ctx := context.Background()
	docker, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return false, errors.Wrap(err, "unable to create new docker client")
	}

	exists, checkErr := docker.RepoImageExists(ctx, t.URL(t.Hash()))
	if checkErr != nil {
		return false, errors.Wrap(err, "CheckRemoteCache#RepoImageExists")
	}
	return exists, nil
}

// PullRemoteCache pulls a docker iamge from a remote repository that matches a target's state hash
func (t Target) PullRemoteCache() error {
	ctx := context.Background()
	docker, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return errors.Wrap(err, "unable to create new docker client")
	}

	return docker.PullImage(ctx, t.URL(t.Hash()))
}

// PushRemoteCache pushes a state artifact to remote storage and pushes a docker image to a remote repository
func (t Target) PushRemoteCache() error {
	ctx := context.Background()

	docker, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return errors.Wrap(err, "unable to create new docker client")
	}
	for _, imageURL := range t.URLsFromTags() {
		if err = docker.PushImage(ctx, imageURL); err != nil {
			return errors.Wrap(err, "an error occurred while pushing the images")
		}
	}
	return nil
}
