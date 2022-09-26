package exec

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubectl/pkg/scheme"
)

func TestKubernetesExecutor(t *testing.T) {
	// FIXME refactor this test to deploy and tear down its own resource for the test
	t.Skip("skipping test because it should not rely on a deployed resource")

	// create a new k8s client
	client := kube.Init(nil)
	getPodTimeout := time.Second * 10
	resourceName := "nginx-ingress"
	currentContext, _ := client.CurrentContext()
	if !utils.IsK8sContextSafe([]string{"development_sre"}, "ARK_K8S_SAFE_CONTEXTS", currentContext) {
		t.Skip("Skipping test because context is not designated as safe")
		return
	}

	// override the default namespace
	client.NamespaceOverride = resourceName

	// query for a pod by its resource type (ds is daemon set)
	pod, err := client.GetPodByResource("ds", resourceName, getPodTimeout)
	require.NoError(t, err)

	// construct an executor against the given pod and run the supplied commands
	err = KubernetesExecutor(KubernetesExecOptions{
		Factory:       client.Factory,
		Pod:           pod,
		Executor:      DefaultKubernetesExecutor(),
		ContainerName: resourceName,
		Command: []string{
			"sh",
			"-c",
			"ls -lha",
		},
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		GetPodTimeout: getPodTimeout,
	})

	require.NoError(t, err)

}

func TestKubernetesExecutorWithMockClient(t *testing.T) {
	version := "v1"
	tests := []struct {
		name, version, podPath, fetchPodPath, execPath string
		pod                                            *corev1.Pod
		execErr                                        bool
	}{
		{
			name:         "pod exec",
			version:      version,
			podPath:      "/api/" + version + "/namespaces/test/pods/foo",
			fetchPodPath: "/namespaces/test/pods/foo",
			execPath:     "/api/" + version + "/namespaces/test/pods/foo/exec",
			pod:          execPod(),
		},
		{
			name:         "pod exec error",
			version:      version,
			podPath:      "/api/" + version + "/namespaces/test/pods/foo",
			fetchPodPath: "/namespaces/test/pods/foo",
			execPath:     "/api/" + version + "/namespaces/test/pods/foo/exec",
			pod:          execPod(),
			execErr:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tf := cmdtesting.NewTestFactory().WithNamespace("test")
			defer tf.Cleanup()

			codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
			ns := scheme.Codecs.WithoutConversion()

			tf.Client = &fake.RESTClient{
				GroupVersion:         schema.GroupVersion{Group: "", Version: "v1"},
				NegotiatedSerializer: ns,
				Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
					switch p, m := req.URL.Path, req.Method; {
					case p == test.podPath && m == "GET":
						body := cmdtesting.ObjBody(codec, test.pod)
						return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: body}, nil
					case p == test.fetchPodPath && m == "GET":
						body := cmdtesting.ObjBody(codec, test.pod)
						return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: body}, nil
					default:
						t.Errorf("%s: unexpected request: %s %#v\n%#v", test.name, req.Method, req.URL, req)
						return nil, fmt.Errorf("unexpected request")
					}
				}),
			}
			tf.ClientConfigVal = &restclient.Config{APIPath: "/api", ContentConfig: restclient.ContentConfig{NegotiatedSerializer: scheme.Codecs, GroupVersion: &schema.GroupVersion{Version: test.version}}}

			params := KubernetesExecOptions{
				Factory:       tf,
				Pod:           execPod(),
				Executor:      FakeKubernetesExecutor(),
				ContainerName: "",
				Command:       []string{"ls"},
			}

			err := KubernetesExecutor(params)
			if !test.execErr && err != nil {
				t.Errorf("%s: Unexpected error: %v", test.name, err)
				return
			}
			if test.execErr {
				return
			}
		})
	}
}

func execPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "test", ResourceVersion: "10"},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			DNSPolicy:     corev1.DNSClusterFirst,
			Containers: []corev1.Container{
				{
					Name: "bar",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}
