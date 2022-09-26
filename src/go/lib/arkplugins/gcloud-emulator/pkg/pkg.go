package pkg

import (
	"bytes"
	"fmt"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type EmulatorOptions struct {
	Name     string `json:"name"`
	Emulator string `json:"emulator"`
	Project  string `json:"project"`
}

func NewGcloudEmulatorManifest(opts EmulatorOptions) (string, error) {
	buff := new(bytes.Buffer)
	manifest := new(objects.Manifest)
	replicas := int32(1)

	cloudSdkContainer := objects.Container(objects.ContainerOptions{
		Name:            "cloud-sdk",
		Image:           "gcr.io/google.com/cloudsdktool/cloud-sdk:latest",
		Command:         []string{"bash", "-c", fmt.Sprintf("gcloud beta emulators %s start --project='%s' --host-port=0.0.0.0:42069", opts.Emulator, opts.Project)},
		ImagePullPolicy: "IfNotPresent",
	})

	podTemp := objects.PodTemplate(objects.PodTemplateOptions{
		Name: "gcloud-emulator",
		Containers: []v1.Container{
			*cloudSdkContainer,
		},
		RestartPolicy: "Always",
		Labels:        map[string]string{"app": opts.Name},
	})

	deployment := objects.Deployment(objects.DeploymentOptions{
		Name:        fmt.Sprintf("%s-emulator", opts.Emulator),
		Replicas:    &replicas,
		PodTemplate: *podTemp,
		Annotations: map[string]string{"app": opts.Name},
	})

	emulatorService := objects.Service(objects.ServiceOptions{
		Name: fmt.Sprintf("%s-service", opts.Emulator),
		Ports: []v1.ServicePort{
			{Name: fmt.Sprintf("%s-service", opts.Emulator),
				Protocol: "TCP",
				Port:     42069,
				TargetPort: intstr.IntOrString{
					IntVal: 42069,
				}},
		},
		Type:     "ClusterIP",
		Selector: map[string]string{"app": opts.Name},
		Labels:   map[string]string{"app": opts.Name},
	})

	manifest.Append(deployment, emulatorService)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
