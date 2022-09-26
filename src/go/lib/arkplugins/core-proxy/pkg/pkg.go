package pkg

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/myfintech/ark/src/go/lib/kube/mutations"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// CoreProxyOptions is all of the things needed to deploy core-proxy
type CoreProxyOptions struct {
	Image              string            `json:"image"`
	Name               string            `json:"name"`
	ServiceAccountName string            `json:"serviceAccountName"`
	Replicas           int32             `json:"replicas"`
	LoadBalancerIP     string            `json:"loadBalancerIp"`
	Environment        string            `json:"environment"`
	EnvVars            map[string]string `json:"envVars"`
	Port               int32             `json:"port"`

	DisableHostNetworking       bool   `json:"disableHostNetworking"`
	DisableOpenTelemetry        bool   `json:"disableOpenTelemetry"`
	DisableGoogleServiceAccount bool   `json:"disableGoogleServiceAccount"`
	DisableVaultSidecar         bool   `json:"disableVaultSidecar"`
	VaultAddr                   string `json:"vaultAddress"`
}

func NewCoreProxyManifest(opts CoreProxyOptions) (string, error) {
	buff := new(bytes.Buffer)
	manifest := new(objects.Manifest)

	vaultConfig := mutations.VaultConfig{
		Team:               "sre",
		App:                "core-proxy",
		Environment:        opts.Environment,
		ClusterEnvironment: "cs",
		Role:               opts.ServiceAccountName,
		DefaultConfig:      "core-proxy/targets",
		Address:            opts.VaultAddr,
	}

	defaultLabels := map[string]string{"app": opts.Name}

	serviceAccountOpts := objects.ServiceAccountOptions{
		Name:   opts.ServiceAccountName,
		Labels: defaultLabels,
	}
	serviceAccount := objects.ServiceAccount(serviceAccountOpts)

	baseEnv := map[string]string{
		"OPEN_TELEMETRY_ENABLED":             strconv.FormatBool(!opts.DisableOpenTelemetry),
		"OPEN_TELEMETRY_STACKDRIVER_ENABLED": "true",
	}

	containerEnv := baseEnv
	containerEnv["PORT"] = strconv.Itoa(int(opts.Port))

	for k, v := range opts.EnvVars {
		containerEnv[k] = v
	}

	formattedContainerEnv := makeEnv(containerEnv)
	formattedInitEnv := makeEnv(opts.EnvVars)

	containerOpts := objects.ContainerOptions{
		Name:    opts.Name,
		Image:   opts.Image,
		Command: []string{"core-proxy", "server"},
		Env:     formattedContainerEnv,
	}

	initContainerOpts := objects.ContainerOptions{
		Name:  fmt.Sprintf("%s-migrations", opts.Name),
		Image: opts.Image,
		Command: []string{
			"/usr/local/bin/monarch",
			"migrate",
			"up",
			"sql/migrations",
		},
		Env: formattedInitEnv,
	}

	if opts.Environment != "dev" {
		envFromSource := []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "core-proxy-config",
					},
				},
			},
		}
		containerOpts.EnvFrom = envFromSource
		initContainerOpts.EnvFrom = envFromSource
	}

	mainContainer := objects.Container(containerOpts)
	initContainer := objects.Container(initContainerOpts)

	podTemplateOpts := objects.PodTemplateOptions{
		Name:                  opts.Name,
		Labels:                defaultLabels,
		Containers:            []corev1.Container{*mainContainer},
		ServiceAccountName:    opts.ServiceAccountName,
		InitContainers:        []corev1.Container{*initContainer},
		HostNetworkingEnabled: !opts.DisableHostNetworking,
	}

	if opts.Environment != "dev" {
		podTemplateOpts.Affinity = corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app",
									Operator: "In",
									Values:   []string{"core-proxy"},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	}

	deploymentOpts := objects.DeploymentOptions{
		Name:        opts.Name,
		Replicas:    &opts.Replicas,
		PodTemplate: *objects.PodTemplate(podTemplateOpts),
	}

	deployment := mutations.ApplyVault(
		!opts.DisableVaultSidecar,
		vaultConfig,
		mutations.ApplyGoogleCloudServiceAccount(
			!opts.DisableGoogleServiceAccount,
			"/etc/google/service_account.json",
			true,
			objects.Deployment(deploymentOpts),
		),
	)

	serviceOpts := objects.ServiceOptions{
		Name: opts.Name,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.IntOrString{IntVal: opts.Port},
			},
		},
		Type:     "ClusterIP",
		Selector: defaultLabels,
		Labels:   defaultLabels,
	}

	service := objects.Service(serviceOpts)

	manifest.Append(serviceAccount, deployment, service)

	if opts.Environment != "dev" {
		internalLBOpts := objects.ServiceOptions{
			Name: fmt.Sprintf("%s-internal-lb", opts.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "tcp",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: opts.Port},
				},
			},
			BackendConfig: "",
			Type:          "LoadBalancer",
			Selector:      defaultLabels,
			Labels:        defaultLabels,
			Annotations: map[string]string{
				"cloud.google.com/load-balancer-type": "Internal",
			},
			LoadBalancerIP: opts.LoadBalancerIP,
		}
		internalService := objects.Service(internalLBOpts)

		manifest.Append(internalService)
	}

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}

func makeEnv(input map[string]string) []corev1.EnvVar {
	env := make([]corev1.EnvVar, 0)
	for k, v := range input {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	return env
}
