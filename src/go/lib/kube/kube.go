package kube

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/utils"
)

// UseLocalConfig creates a kubernetes clientset based on the local configurations
// Deprecated
// we are moving toward using the kubectl factory functions

// UseInClusterConfig assumes its caller resides in a kubernetes cluster and attempts to load its configuration from the cluster API
// Deprecated
// we are moving toward using the kubectl factory functions

// Client is a container for a kubernetes cmdutil factory
type Client struct {
	Factory           cmdutil.Factory
	NamespaceOverride string
	OutputWriter      io.Writer
}

func stringptr(s string) *string {
	return &s
}

func configFromEnv() *string {
	config := os.Getenv("KUBECONFIG")
	if config != "" {
		return stringptr(config)
	} else {
		return nil
	}
}

// Init creates a new MANTL k8s kube.Client
// This will attempt to match the server API version
func Init(getter genericclioptions.RESTClientGetter) Client {
	if getter == nil {
		config := genericclioptions.NewConfigFlags(true)
		config.KubeConfig = configFromEnv()
		getter = cmdutil.NewMatchVersionFlags(config)
	}

	return Client{
		Factory:      cmdutil.NewFactory(getter),
		OutputWriter: os.Stderr,
	}
}

// Namespace returns the current clients namespace
// - Checks namespace override
// - Checks kube config namespace
// - Defaults to v1.NamespaceDefault
func (c *Client) Namespace() string {
	if c.NamespaceOverride != "" {
		return c.NamespaceOverride
	}
	if ns, _, err := c.Factory.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return v1.NamespaceDefault
}

// CurrentContext returns the current context from the kube config
func (c *Client) CurrentContext() (string, error) {
	config, err := c.Factory.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return "", err
	}
	return config.CurrentContext, nil
}

// NewNamespacedBuilder returns a new resource builder for structured api objects.
func (c *Client) NewNamespacedBuilder() *resource.Builder {
	return c.Factory.NewBuilder().
		ContinueOnError().
		NamespaceParam(c.Namespace()).
		DefaultNamespace().
		Flatten()
}

// GetPodByResource queries a kubernetes cluster for a corev1.pod by deployment
func (c *Client) GetPodByResource(
	resourceType, resourceName string,
	timeout time.Duration,
) (*corev1.Pod, error) {
	formattedResourceName := fmt.Sprintf("%s/%s", resourceType, resourceName)
	builder := c.NewNamespacedBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ResourceNames("pods", formattedResourceName)

	object, err := builder.Do().Object()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get pods from %s", formattedResourceName)
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return polymorphichelpers.AttachablePodForObjectFn(c.Factory, object, timeout)
}

// InitWithSafeContexts initializes a new kubernetes client set
func InitWithSafeContexts(namespace string, safeContexts []string) (Client, error) {
	K8s := Init(nil)
	safe := false
	currentContext, err := K8s.CurrentContext()
	if err != nil {
		return K8s, err
	}
	if currentContext == "" {
		return K8s, nil
	}
	for _, safeContext := range safeContexts {
		if safe = safeContext == currentContext; safe {
			break
		}
	}
	if !safe {
		return K8s, errors.Errorf("Your current k8s context (%s) is unsafe, must be one of:\n%s", currentContext, strings.Join(safeContexts, ", \n"))
	}
	if namespace != "" {
		K8s.NamespaceOverride = namespace
	}

	return K8s, nil
}

// DefaultSafeContexts returns known safe Kubernetes contexts
func DefaultSafeContexts() []string {
	return []string{
		"local",
		"docker-desktop",
		"docker-for-desktop",
		"minikube",
		"kind",
	}
}
