package mutations

import (
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
)

// VaultConfig represents the fields necessary to have a service authenticate with and pull secrets from Vault
type VaultConfig struct {
	Team               string `json:"vaultTeam"`
	App                string `json:"vaultApp"`
	Environment        string `json:"vaultEnv"`
	ClusterEnvironment string `json:"clusterEnv"`
	Role               string `json:"vaultRole"`
	DefaultConfig      string `json:"vaultDefaultConfig"`
	Address            string `json:"vaultAddress"`
}

// ApplyVault adds configuration information - environment variables, volume mounts, containers, and init containers to a deployment object
func ApplyVault(enableVault bool, vaultConfig VaultConfig, input *v1.Deployment) *v1.Deployment {
	if enableVault {
		return ApplyVaultConfigToDeployment(vaultConfig, ApplyVaultContainers(input))
	}
	return input
}

// ApplyVaultContainers appends the vault-login init container and the vault-auto-renew container to the deployment object
func ApplyVaultContainers(input *v1.Deployment) *v1.Deployment {
	vaultImage := "gcr.io/[insert-google-project]/domain/vault-auth:f208f4e"
	input.Spec.Template.Spec.InitContainers = append(input.Spec.Template.Spec.InitContainers, v12.Container{
		Name:    "vault-login",
		Image:   vaultImage,
		Command: []string{"/usr/local/bin/vault-auth", "k8s-login"},
	})
	input.Spec.Template.Spec.Containers = append(input.Spec.Template.Spec.Containers, v12.Container{
		Name:    "vault-auto-renew",
		Image:   vaultImage,
		Command: []string{"/usr/local/bin/vault-auth", "renew"},
		Lifecycle: &v12.Lifecycle{
			PreStop: &v12.Handler{
				Exec: &v12.ExecAction{
					Command: []string{"/usr/local/bin/vault-auth", "revoke"},
				},
			},
		},
	})
	return input

}

// ApplyVaultConfigToDeployment appends a set of environment variables, a set of annotations, and a volume/volume mount to all containers in a deployment object
func ApplyVaultConfigToDeployment(vaultConfig VaultConfig, input *v1.Deployment) *v1.Deployment {
	containerUpdate := func(container *v12.Container) *v12.Container {
		container.Env = append(container.Env, []v12.EnvVar{
			{
				Name:  "AUTH_SERVICE_ADDR",
				Value: "http://auth-service.vault:5000",
			},
			{
				Name:  "CLUSTER_ENV",
				Value: vaultConfig.ClusterEnvironment,
			},
			{
				Name:  "VAULT_CONFIG",
				Value: "/etc/vault/config.json",
			},
			{
				Name:  "VAULT_ADDR",
				Value: vaultConfig.Address,
			},
			{
				Name:  "VAULT_TEAM",
				Value: vaultConfig.Team,
			},
			{
				Name:  "VAULT_ENV",
				Value: vaultConfig.Environment,
			},
			{
				Name:  "VAULT_APP",
				Value: vaultConfig.App,
			},
			{
				Name:  "VAULT_ROLE",
				Value: vaultConfig.Role,
			},
			{
				Name:  "TOKEN_TTL_INCREMENT",
				Value: "86400",
			},
			{
				Name:  "VAULT_DEFAULT_CONFIG",
				Value: vaultConfig.DefaultConfig,
			},
			{
				Name: "POD_NAMESPACE",
				ValueFrom: &v12.EnvVarSource{
					FieldRef: &v12.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		}...)
		container.VolumeMounts = append(container.VolumeMounts, v12.VolumeMount{
			Name:      "vault-config",
			MountPath: "/etc/vault",
		})
		return container
	}
	annotations := map[string]string{
		"vault.app":  vaultConfig.App,
		"vault.team": vaultConfig.Team,
	}
	if input.Annotations == nil {
		input.SetAnnotations(annotations)
	} else {
		for k, v := range annotations {
			input.Annotations[k] = v
		}
	}
	if input.Spec.Template.Annotations == nil {
		input.Spec.Template.SetAnnotations(annotations)
	} else {
		for k, v := range annotations {
			input.Spec.Template.ObjectMeta.Annotations[k] = v
		}
	}
	for i, container := range input.Spec.Template.Spec.Containers {
		input.Spec.Template.Spec.Containers[i] = *containerUpdate(&container)
	}
	for i, initContainer := range input.Spec.Template.Spec.InitContainers {
		input.Spec.Template.Spec.InitContainers[i] = *containerUpdate(&initContainer)
	}
	input.Spec.Template.Spec.Volumes = append(input.Spec.Template.Spec.Volumes, v12.Volume{
		Name: "vault-config",
		VolumeSource: v12.VolumeSource{
			EmptyDir: &v12.EmptyDirVolumeSource{},
		},
	})
	return input
}

// ApplyGoogleCloudServiceAccount appends all containers with the GOOGLE_APPLICATION_CREDENTIALS env var as well as the appropriate volume/volume mount
func ApplyGoogleCloudServiceAccount(enableGCSA bool, filename string, mountAsSecret bool, input *v1.Deployment) *v1.Deployment {
	if enableGCSA {
		containerUpdate := func(container *v12.Container) *v12.Container {
			container.Env = append(container.Env, v12.EnvVar{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: filename,
			})
			container.VolumeMounts = append(container.VolumeMounts, v12.VolumeMount{
				Name:      "google-service-account",
				MountPath: "/etc/google",
			})
			return container
		}
		for i, container := range input.Spec.Template.Spec.Containers {
			input.Spec.Template.Spec.Containers[i] = *containerUpdate(&container)
		}
		for i, initContainer := range input.Spec.Template.Spec.InitContainers {
			input.Spec.Template.Spec.InitContainers[i] = *containerUpdate(&initContainer)
		}
		volumeSource := v12.VolumeSource{}
		if mountAsSecret {
			volumeSource = v12.VolumeSource{
				Secret: &v12.SecretVolumeSource{
					SecretName: "google-service-account",
				},
			}
		} else {
			volumeSource = v12.VolumeSource{
				EmptyDir: &v12.EmptyDirVolumeSource{},
			}
		}
		input.Spec.Template.Spec.Volumes = append(input.Spec.Template.Spec.Volumes, v12.Volume{
			Name:         "google-service-account",
			VolumeSource: volumeSource,
		})
		return input
	}
	return input
}

// ApplySingleHostAffinity sets the pod template to use single host affinity for the deployment

// ApplyLivenessAndReadinessProbes applies liveness and readiness probes to containers

// ApplyResourceConstraints applies resource constraints to containers

// ApplyImagePullPolicy applies an image pull policy, overriding the value that was originally set on the container object(s)
