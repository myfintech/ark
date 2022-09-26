package pkg

import (
	"bytes"

	"github.com/myfintech/ark/src/go/lib/kube/mutations"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type CiCdOptions struct {
	Image              string `json:"image"`
	Name               string `json:"name"`
	ServiceAccountName string `json:"serviceAccountName"`
	Replicas           int32
}

func NewCiCdGhaRunnersManifest(opts CiCdOptions) (string, error) {
	buff := new(bytes.Buffer)
	manifest := new(objects.Manifest)

	vaultConfig := mutations.VaultConfig{
		Team:               "sre",
		App:                "actions-runner",
		Environment:        "es",
		ClusterEnvironment: "cicd",
		Role:               "actions-runner-es",
		DefaultConfig:      "actions-runner/default_config",
		Address:            "http://vault.es.mantl.internal",
	}

	defaultLabels := map[string]string{"app": opts.Name}

	serviceAccountOpts := objects.ServiceAccountOptions{
		Name:   opts.ServiceAccountName,
		Labels: defaultLabels,
	}

	serviceAccount := objects.ServiceAccount(serviceAccountOpts)

	ghaContainerEnv := makeEnv(map[string]string{
		"DOCKER_TLS_CERTDIR": "/docker/certs",
		"DOCKER_TLS_VERIFY":  "1",
		"DOCKER_HIST":        "127.0.0.1:2376",
		"DOCKER_PORT":        "2376",
		"DOCKER_CERT_PATH":   "/docker/certs/client",
		"ARK_DATA_HOME":      "/ark",
	})

	containerVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "ark-artifacts",
			MountPath: "/ark",
		},
		{
			Name:      "docker-tls-certs",
			MountPath: "/docker/certs",
		},
		{
			Name:      "runner-work-dir",
			MountPath: "/runner",
		},
		{
			Name:      "watchman-socket",
			MountPath: "/usr/local/var/run/watchman",
		},
	}

	privileged := true

	ghaContainer := objects.Container(objects.ContainerOptions{
		Name:         "actions-runner",
		Image:        opts.Image,
		Env:          ghaContainerEnv,
		VolumeMounts: containerVolumeMounts,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"memory": resource.MustParse("8Gi"),
			},
			Requests: corev1.ResourceList{
				"memory": resource.MustParse("8Gi"),
			},
		},
		Lifecycle: corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"/bin/bash",
						"-c",
						"RUNNER_ALLOW_RUNASROOT=1 ./config.sh remove --token $(curl -sS --request POST --url \"https://api.github.com/repos/myfintech/mantl/actions/runners/remove-token\" --header \"authorization: Bearer ${GITHUB_API_TOKEN}\"  --header \"content-type: application/json\" | jq -r .token)",
					},
				},
			},
		},
	})

	dindContainerVolumeMounts := append(containerVolumeMounts, corev1.VolumeMount{
		Name:      "root-volume",
		MountPath: "/root",
	})

	dindContainer := objects.Container(objects.ContainerOptions{
		Name:  "dockerd",
		Image: "docker:dind",
		Env: []corev1.EnvVar{
			{
				Name:  "DOCKER_TLS_CERTDIR",
				Value: "/docker/certs",
			},
		},
		VolumeMounts: dindContainerVolumeMounts,
		SecurityContext: corev1.SecurityContext{
			Privileged: &privileged,
		},
	})

	templateVolumes := []corev1.Volume{
		{
			Name: "ark-artifacts",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "docker-tls-certs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "runner-work-dir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "root-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "watchman-socket",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	podTemplateOpts := objects.PodTemplateOptions{
		Name: opts.Name,
		Containers: []corev1.Container{
			*ghaContainer,
			*dindContainer,
		},
		ServiceAccountName: opts.ServiceAccountName,
		Volumes:            templateVolumes,
		Labels:             defaultLabels,
	}

	deploymentOpts := objects.DeploymentOptions{
		Name:        opts.Name,
		Replicas:    &opts.Replicas,
		PodTemplate: *objects.PodTemplate(podTemplateOpts),
	}

	deployment := mutations.ApplyVault(true, vaultConfig, mutations.ApplyGoogleCloudServiceAccount(
		true,
		"/etc/google/service_account.json",
		true,
		objects.Deployment(deploymentOpts),
	))

	manifest.Append(serviceAccount, deployment)

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

	env = append(env, corev1.EnvVar{
		Name: "GITHUB_API_TOKEN",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "github-api-token",
				},
				Key: "token",
			},
		},
	})
	return env
}
