package pkg

import (
	"bytes"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewSDMManifest() (string, error) {
	manifest := new(objects.Manifest)
	buff := new(bytes.Buffer)

	name := "sdm"
	replicas := int32(1)
	secretBool := true

	sdmContainer := objects.Container(objects.ContainerOptions{
		Name:    "sdm",
		Image:   "gcr.io/managed-infrastructure/mantl/sdm-ark:latest",
		Command: []string{"bash", "-c", "/run.sh"},
		Env: []v1.EnvVar{{
			Name: "SDM_SERVICE_TOKEN",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					Key: "data",
					LocalObjectReference: v1.LocalObjectReference{
						Name: "sdm-service-account",
					},
					Optional: &secretBool,
				},
			},
		},
		},
		VolumeMounts: []v1.VolumeMount{{
			Name:      "sdm",
			ReadOnly:  false,
			MountPath: "/tmp/sdm",
		},
		},
		ImagePullPolicy: "IfNotPresent",
	})

	nginxContainer := objects.Container(objects.ContainerOptions{
		Name:            "nginx",
		Image:           "gcr.io/managed-infrastructure/mantl/nginx-ark:latest",
		Env:             []v1.EnvVar{},
		ImagePullPolicy: "IfNotPresent",
	})

	podTemp := objects.PodTemplate(objects.PodTemplateOptions{
		Name: "sdm",
		Containers: []v1.Container{
			*sdmContainer,
			*nginxContainer,
		},
		RestartPolicy: "Always",
		Volumes: []v1.Volume{{
			Name: "sdm",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "sdm",
					Optional:   &secretBool,
				},
			},
		}},
		Labels: map[string]string{"app": name},
	})

	deployment := objects.Deployment(objects.DeploymentOptions{
		Name:        "sdm",
		Replicas:    &replicas,
		PodTemplate: *podTemp,
		Annotations: map[string]string{"app": name},
	})

	vaultService := objects.Service(objects.ServiceOptions{
		Name: "vault-sdm",
		Ports: []v1.ServicePort{
			{Name: "vault-sdm",
				Protocol: "TCP",
				Port:     80,
				TargetPort: intstr.IntOrString{
					IntVal: 8080,
				}},
		},
		Type:     "ClusterIP",
		Selector: map[string]string{"app": name},
		Labels:   map[string]string{"app": name},
	})

	coreProxyService := objects.Service(objects.ServiceOptions{
		Name: "core-proxy-int-sdm",
		Ports: []v1.ServicePort{
			{Name: "core-proxy-int-sdm",
				Protocol: "TCP",
				Port:     80,
				TargetPort: intstr.IntOrString{
					IntVal: 8081,
				}},
		},
		Type:     "ClusterIP",
		Selector: map[string]string{"app": name},
		Labels:   map[string]string{"app": name},
	})

	manifest.Append(deployment, vaultService, coreProxyService)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
