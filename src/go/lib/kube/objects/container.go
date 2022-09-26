package objects

import corev1 "k8s.io/api/core/v1"

// ContainerOptions represents fields that can be passed in to create a container
type ContainerOptions struct {
	Name            string
	Image           string
	Command         []string
	Env             []corev1.EnvVar
	Args            []string
	Ports           []corev1.ContainerPort
	EnvFrom         []corev1.EnvFromSource
	VolumeMounts    []corev1.VolumeMount
	Resources       corev1.ResourceRequirements
	ImagePullPolicy corev1.PullPolicy
	SecurityContext corev1.SecurityContext
	Lifecycle       corev1.Lifecycle
	LivenessProbe   corev1.Probe
	ReadinessProbe  corev1.Probe
	WorkingDir      string
}

// Container returns a pointer to a container object
func Container(opts ContainerOptions) *corev1.Container {
	if opts.ImagePullPolicy == "" {
		opts.ImagePullPolicy = "IfNotPresent"
	}

	container := &corev1.Container{
		Name:            opts.Name,
		Image:           opts.Image,
		Command:         opts.Command,
		Args:            opts.Args,
		Ports:           opts.Ports,
		EnvFrom:         opts.EnvFrom,
		Env:             opts.Env,
		Resources:       opts.Resources,
		VolumeMounts:    opts.VolumeMounts,
		ImagePullPolicy: opts.ImagePullPolicy,
		SecurityContext: &opts.SecurityContext,
		WorkingDir:      opts.WorkingDir,
	}

	// an error gets thrown if empty lifecycles/probes are on the resource, so we only want them to be on the resource if they have been defined
	if opts.Lifecycle != (corev1.Lifecycle{}) {
		container.Lifecycle = &opts.Lifecycle
	}
	if opts.LivenessProbe != (corev1.Probe{}) {
		container.LivenessProbe = &opts.LivenessProbe
	}
	if opts.ReadinessProbe != (corev1.Probe{}) {
		container.ReadinessProbe = &opts.ReadinessProbe
	}

	return container
}
