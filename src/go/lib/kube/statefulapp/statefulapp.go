package statefulapp

import (
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/kube/mutations"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Options represents all of the input we accept in our Jsonnet microservice implementation
type Options struct {
	mutations.VaultConfig

	Replicas                        int32                   `json:"replicas"`
	Port                            int32                   `json:"port,omitempty"`
	ServicePort                     int32                   `json:"servicePort,omitempty"`
	Name                            string                  `json:"name"`
	Image                           string                  `json:"image"`
	ServiceType                     string                  `json:"serviceType,omitempty"`
	Command                         []string                `json:"command,omitempty"`
	HostNetwork                     bool                    `json:"hostNetwork"`
	Env                             map[string]string       `json:"env,omitempty"`
	EnableVault                     bool                    `json:"enableVault"`
	EnableGoogleCloudServiceAccount bool                    `json:"enableGoogleCloudServiceAccount"`
	DataDir                         string                  `json:"dataDir"`
	Capacity                        string                  `json:"capacity"`
	Ingress                         *objects.IngressOptions `json:"ingress,omitempty"`
	ReadinessProbe                  *objects.ProbeOptions   `json:"readinessProbe,omitempty"`
	LivenessProbe                   *objects.ProbeOptions   `json:"livenessProbe,omitempty"`
}

// NewStatefulApp returns a pointer to a complete Kubernetes manifest for a stateful set
func NewStatefulApp(opts Options) *objects.Manifest {
	manifest := new(objects.Manifest)
	defaultLabels := map[string]string{"app": opts.Name}

	if opts.Capacity == "" {
		opts.Capacity = "5Gi"
	}

	env := make([]v1.EnvVar, 0)
	for k, v := range opts.Env {
		env = append(env, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	pvcOpts := objects.PersistentVolumeOptions{
		Name:    opts.Name,
		Storage: opts.Capacity,
	}

	pvcs := objects.BuildVolumeClaimsTemplate([]objects.PersistentVolumeOptions{pvcOpts})

	dataDir := opts.Name
	if opts.DataDir != "" {
		dataDir = opts.DataDir
	}

	volumeMounts := make([]v1.VolumeMount, 0)
	for _, pvc := range *pvcs {
		mountPath := filepath.Join("/mnt/", dataDir)
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      pvc.Name,
			MountPath: mountPath,
		})
	}

	containerOpts := objects.ContainerOptions{
		Name:         opts.Name,
		Image:        opts.Image,
		Command:      opts.Command,
		Env:          env,
		Resources:    v1.ResourceRequirements{},
		VolumeMounts: volumeMounts,
	}

	if opts.ReadinessProbe != nil {
		containerOpts.ReadinessProbe = objects.Probe(*opts.ReadinessProbe)
	}

	if opts.LivenessProbe != nil {
		containerOpts.LivenessProbe = objects.Probe(*opts.LivenessProbe)
	}

	podTemplateOpts := objects.PodTemplateOptions{
		Name:                  opts.Name,
		Containers:            []v1.Container{*objects.Container(containerOpts)},
		HostNetworkingEnabled: opts.HostNetwork,
		Labels:                defaultLabels,
	}

	statefulSetOpts := objects.StatefulSetOptions{
		Name:        opts.Name,
		Replicas:    opts.Replicas,
		PodTemplate: *objects.PodTemplate(podTemplateOpts),
		PVCs:        *pvcs,
	}

	statefulSet := objects.StatefulSet(statefulSetOpts)

	serviceOpts := objects.ServiceOptions{
		Name: opts.Name,
		Ports: []v1.ServicePort{
			{
				// FIXME: the port name will not always be 'http'
				// K8s has implied behavior with named ports
				// https://kubernetes.io/docs/concepts/services-networking/service/#dns
				Name:       "http",
				Port:       opts.ServicePort,
				TargetPort: intstr.IntOrString{IntVal: opts.Port},
			},
		},
		Selector: defaultLabels,
		Type:     v1.ServiceType(opts.ServiceType),
	}

	service := objects.Service(serviceOpts)

	manifest.Append(statefulSet, service)
	if opts.Ingress != nil {
		ingress := objects.Ingress(*opts.Ingress)
		manifest.Append(ingress)
	}
	return manifest
}
