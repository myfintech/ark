package pkg

import (
	"bytes"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
)

type TerraformCloudAgentOptions struct {
	Name       string `json:"name"`
	AgentToken string `json:"agentToken"`
	Replicas   int32  `json:"replicas"`
}

func NewTerraformCloudAgentManifest(opts TerraformCloudAgentOptions) (string, error) {
	manifest := new(objects.Manifest)
	buff := new(bytes.Buffer)

	tfAgentContainer := objects.Container(objects.ContainerOptions{
		Name:  opts.Name,
		Image: "hashicorp/tfc-agent:latest",
		Env: []corev1.EnvVar{
			{Name: "TFC_AGENT_TOKEN", Value: opts.AgentToken},
		},
	})

	podTemplate := objects.PodTemplate(objects.PodTemplateOptions{
		Name:          opts.Name,
		Containers:    []corev1.Container{*tfAgentContainer},
		RestartPolicy: "Always",
		Labels: map[string]string{
			"app": opts.Name,
		},
	})

	deployment := objects.Deployment(objects.DeploymentOptions{
		Name:        opts.Name,
		Replicas:    &opts.Replicas,
		PodTemplate: *podTemplate,
	})

	manifest.Append(deployment)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
