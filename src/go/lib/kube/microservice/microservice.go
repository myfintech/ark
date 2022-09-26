package microservice

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/myfintech/ark/src/go/lib/kube/mutations"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
)

// Options represents all the input we accept in our Jsonnet microservice implementation
type Options struct {
	mutations.VaultConfig

	Replicas                        int32                   `json:"replicas"`
	Port                            int32                   `json:"port,omitempty"`
	ServicePort                     int32                   `json:"servicePort,omitempty"`
	Name                            string                  `json:"name"`
	ServiceAccountName              string                  `json:"serviceAccountName"`
	Image                           string                  `json:"image"`
	ServiceType                     string                  `json:"serviceType,omitempty"`
	Command                         []string                `json:"command,omitempty"`
	Env                             map[string]string       `json:"env,omitempty"`
	EnableVault                     bool                    `json:"enableVault"`
	EnableHostNetworking            bool                    `json:"hostNetwork"`
	EnableGoogleCloudServiceAccount bool                    `json:"enableGoogleCloudServiceAccount"`
	IncludeKubeletHostIp            bool                    `json:"includeKubeletHostIp"`
	ReadinessProbe                  *objects.ProbeOptions   `json:"readinessProbe,omitempty"`
	LivenessProbe                   *objects.ProbeOptions   `json:"livenessProbe,omitempty"`
	Ingress                         *objects.IngressOptions `json:"ingress,omitempty"`
}

// NewMicroService returns a pointer to a complete Kubernetes manifest for a deployment
func NewMicroService(opts Options) *objects.Manifest {
	manifest := new(objects.Manifest)
	defaultLabels := map[string]string{"app": opts.Name}

	env := make([]v1.EnvVar, 0)
	for k, v := range opts.Env {
		env = append(env, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	if opts.IncludeKubeletHostIp {
		env = append(env, v1.EnvVar{
			Name: "KUBELET_HOST_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		})
	}

	serviceAccountOpts := objects.ServiceAccountOptions{
		Name:   opts.ServiceAccountName,
		Labels: defaultLabels,
	}

	serviceAccount := objects.ServiceAccount(serviceAccountOpts)

	containerOpts := objects.ContainerOptions{
		Name:      opts.Name,
		Image:     opts.Image,
		Command:   opts.Command,
		Env:       env,
		Resources: v1.ResourceRequirements{},
	}

	if opts.ReadinessProbe != nil {
		containerOpts.ReadinessProbe = objects.Probe(*opts.ReadinessProbe)
	}

	if opts.LivenessProbe != nil {
		containerOpts.LivenessProbe = objects.Probe(*opts.LivenessProbe)
	}

	podTemplateOpts := objects.PodTemplateOptions{
		Name:                  opts.Name,
		ServiceAccountName:    serviceAccount.Name,
		Containers:            []v1.Container{*objects.Container(containerOpts)},
		HostNetworkingEnabled: opts.EnableHostNetworking,
		Labels:                defaultLabels,
	}

	deploymentOpts := objects.DeploymentOptions{
		Name:        opts.Name,
		Replicas:    &opts.Replicas,
		PodTemplate: *objects.PodTemplate(podTemplateOpts),
	}

	deployment := mutations.ApplyGoogleCloudServiceAccount(
		opts.EnableGoogleCloudServiceAccount,
		"/etc/google/service_account.json",
		false,
		mutations.ApplyVault(opts.EnableVault, opts.VaultConfig, objects.Deployment(deploymentOpts)))

	manifest.Append(serviceAccount, deployment)

	if opts.ServicePort != 0 {
		serviceOpts := objects.ServiceOptions{
			Name: opts.Name,
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       opts.ServicePort,
					TargetPort: intstr.IntOrString{IntVal: opts.Port},
				},
			},
			Selector: defaultLabels,
			Type:     v1.ServiceType(opts.ServiceType),
		}

		service := objects.Service(serviceOpts)
		manifest.Append(service)
	}

	if opts.Ingress != nil {
		ingress := objects.Ingress(*opts.Ingress)
		manifest.Append(ingress)
	}
	return manifest
}
