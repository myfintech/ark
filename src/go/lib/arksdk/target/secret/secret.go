package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/kube"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/hashicorp/hcl/v2"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/zclconf/go-cty/cty"
)

// Target defines the required and optional attributes for defining a secret
type Target struct {
	*base.RawTarget `json:"-"`

	Optional    hcl.Expression `hcl:"optional,attr"`
	SecretName  hcl.Expression `hcl:"secret_name,attr"`
	Namespace   hcl.Expression `hcl:"namespace,optional"`
	Files       hcl.Expression `hcl:"files,optional"`
	Environment hcl.Expression `hcl:"environment,optional"`
}

// ComputedAttrs is used to store the computed values from the Target expressions
type ComputedAttrs struct {
	Optional    bool     `hcl:"optional,attr"`
	SecretName  string   `hcl:"secret_name,attr"`
	Namespace   string   `hcl:"namespace,optional"`
	Files       []string `hcl:"files,optional"`
	Environment []string `hcl:"environment,optional"`
}

// Attributes return a combined map of rawTarget.Attributes and typedTarget.Attributes
func (t Target) Attributes() map[string]cty.Value {
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{})
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

// Build creates or updates a kubernetes secret
func (t Target) Build() error {
	attrs := t.ComputedAttrs()

	if err := t.MkArtifactsDir(); err != nil {
		return err
	}

	client := t.Workspace.K8s
	namespace := client.Namespace()
	if attrs.Namespace != "" {
		namespace = attrs.Namespace
	}

	if (attrs.Files == nil && attrs.Environment == nil) || (attrs.Files != nil && attrs.Environment != nil) {
		return errors.New("only one of 'files' or 'environment' must be provided")
	}

	secretData := make(map[string]string, 0)

	if attrs.Files != nil {
		data, err := t.processFilesForSecretData()
		if err != nil {
			return err
		}
		secretData = data
	}

	if attrs.Environment != nil {
		data, err := t.processEnvForSecretData()
		if err != nil {
			return err
		}
		secretData = data
	}

	// if there is no data to write to the secret and all of the values are optional, don't write a secret
	// if a value wasn't optional, an error would be thrown in an earlier function call
	if len(secretData) == 0 && attrs.Optional {
		log.Warn("no valid files or env vars have been provided; not creating secret")
		return nil
	}

	labels := map[string]string{
		"target.hash": t.ShortHash(), // have to use short hash because regular hash is too long for label value
	}

	_, err := kube.CreateOrUpdateSecret(client, namespace, attrs.SecretName, secretData, labels)
	return err
}

// CheckLocalBuildCache checks the configured K8s context to see if a secret with the same target hash already exists in the cluster
func (t Target) CheckLocalBuildCache() (bool, error) {
	restClient, err := t.Workspace.K8s.Factory.RESTClient()
	if err != nil {
		return false, err
	}
	namespace := t.Workspace.K8s.Namespace()
	if t.ComputedAttrs().Namespace != "" {
		namespace = t.ComputedAttrs().Namespace
	}
	secret, err := kube.GetSecret(restClient, namespace, t.ComputedAttrs().SecretName)
	if err != nil {
		return false, err
	}
	if secret != nil && secret.Labels["target.hash"] == t.ShortHash() {
		return true, nil
	}
	return false, nil
}

// CheckRemoteCache overrides the RawTarget interface method because it wouldn't work for this target
func (t Target) CheckRemoteCache() (bool, error) {
	return false, nil
}

// PullRemoteCache overrides the RawTarget interface method because it wouldn't work for this target
func (t Target) PullRemoteCache() error {
	return nil
}

// PushRemoteCache overrides the RawTarget interface method because it wouldn't work for this target
func (t Target) PushRemoteCache() error {
	return nil
}

func (t Target) processFilesForSecretData() (map[string]string, error) {
	secretData := make(map[string]string, 0)
	attrs := t.ComputedAttrs()
	for _, path := range attrs.Files {
		if !filepath.IsAbs(path) {
			absPath := filepath.Join(t.Dir, path)

			path = absPath
		}
		info, err := os.Stat(path)
		if err != nil && !attrs.Optional {
			return secretData, err
		} else if err != nil && attrs.Optional {
			log.Warnf("there's an issue with the path '%s', but it's marked as optional; skipping", path)
			fmt.Println(err)
			continue
		}
		if info != nil && info.IsDir() {
			data, processDirErr := processDirForSecretData(path)
			if processDirErr != nil {
				return secretData, processDirErr
			}
			for k, v := range data {
				secretData[k] = v
			}
			continue
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return secretData, err
		}
		secretData[info.Name()] = string(content)
	}
	return secretData, nil
}

func processDirForSecretData(path string) (map[string]string, error) {
	secretData := make(map[string]string, 0)
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return secretData, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			log.Warnf("%s is a directory, and recursion is not supported; skipping", entry.Name())
			continue
		}
		subPath := filepath.Join(path, entry.Name())
		content, readFileErr := ioutil.ReadFile(subPath)
		if readFileErr != nil {
			return secretData, readFileErr
		}
		secretData[entry.Name()] = string(content)
	}
	return secretData, nil
}

func (t Target) processEnvForSecretData() (map[string]string, error) {
	secretData := make(map[string]string, 0)
	attrs := t.ComputedAttrs()
	for _, envVar := range attrs.Environment {
		value, exists := os.LookupEnv(envVar)
		if !exists && !attrs.Optional {
			return secretData, errors.New(fmt.Sprintf("env var '%s' does not exist and is not optional", envVar))
		} else if !exists && attrs.Optional {
			log.Warnf("env var '%s' does not exist, but is marked as optional; skipping", envVar)
			continue
		}
		secretData[envVar] = value
	}
	return secretData, nil
}
