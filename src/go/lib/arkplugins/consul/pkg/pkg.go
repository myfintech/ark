package pkg

import (
	"bytes"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ConsulOptions struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
}

func NewConsulManifest(opts ConsulOptions) (string, error) {
	buff := new(bytes.Buffer)
	manifest := new(objects.Manifest)

	clientServiceAccountOpts := objects.ServiceAccountOptions{
		Name:   opts.Name + "-client",
		Labels: buildLabels(nil),
	}

	serverServiceAccountOpts := objects.ServiceAccountOptions{
		Name:   opts.Name + "-server",
		Labels: buildLabels(nil),
	}

	disruptionBudgetLabels := buildLabels(map[string]string{"component": "client"})

	podDisruptionBudgetOpts := objects.PodDisruptionBudgetOptions{
		Name:           opts.Name + "server",
		MaxUnavailable: 1,
		Labels:         buildLabels(nil),
		MatchLabels:    disruptionBudgetLabels,
	}

	clientRoleOpts := objects.RoleOptions{
		Name:   opts.Name + "-client",
		Labels: buildLabels(nil),
		Rules:  []rbacv1.PolicyRule{},
	}

	serverRoleOpts := objects.RoleOptions{
		Name:   opts.Name + "-server",
		Labels: buildLabels(nil),
		Rules:  []rbacv1.PolicyRule{},
	}

	clientConfigMapOpts := objects.ConfigMapOptions{
		Name:   opts.Name + "-client-config",
		Labels: buildLabels(nil),
		Data: map[string]string{
			"central-config.json":    `{"enable_central_service_config": true}`,
			"extra-from-values.json": "{}",
		},
	}

	serverConfigMapOpts := objects.ConfigMapOptions{
		Name:   opts.Name + "-server-config",
		Labels: buildLabels(nil),
		Data: map[string]string{
			"central-config.json":    `{"enable_central_service_config": true}`,
			"extra-from-values.json": "{}",
		},
	}

	clientRoleBindingOpts := objects.RoleBindingOptions{
		Name: opts.Name + "-client",
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: opts.Name + "-client",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     opts.Name + "-client",
		},
		Labels: buildLabels(nil),
	}

	serverRoleBindingOpts := objects.RoleBindingOptions{
		Name: opts.Name + "-server",
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: opts.Name + "-server",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     opts.Name + "-server",
		},
		Labels: buildLabels(nil),
	}

	dnsServiceLabels := buildLabels(map[string]string{"component": "dns"})
	dnsServiceSelector := buildLabels(map[string]string{"hasDNS": "true"})

	dnsServiceOpts := objects.ServiceOptions{
		Name: opts.Name + "-dns",
		Ports: []v1.ServicePort{
			{
				Name:       "dns-tcp",
				Protocol:   "TCP",
				Port:       53,
				TargetPort: intstr.Parse("dns-tcp"),
			},
			{
				Name:       "dns-udp",
				Protocol:   "UDP",
				Port:       53,
				TargetPort: intstr.Parse("dns-udp"),
			},
		},
		Type:     "ClusterIP",
		Selector: dnsServiceSelector,
		Labels:   dnsServiceLabels,
	}

	headlessDNSServiceLabels := buildLabels(map[string]string{"component": "server"})

	headlessDNSServiceOpts := objects.ServiceOptions{
		Name: opts.Name + "-server",
		Ports: []v1.ServicePort{
			{
				Name:       "http",
				Port:       8500,
				TargetPort: intstr.FromInt(8500),
			},
			{
				Name:       "serflan-tcp",
				Protocol:   "TCP",
				Port:       8301,
				TargetPort: intstr.FromInt(8301),
			},
			{
				Name:       "serflan-udp",
				Protocol:   "UDP",
				Port:       8301,
				TargetPort: intstr.FromInt(8301),
			},
			{
				Name:       "serfwan-tcp",
				Protocol:   "TCP",
				Port:       8302,
				TargetPort: intstr.FromInt(8302),
			},
			{
				Name:       "serfwan-udp",
				Protocol:   "UDP",
				Port:       8302,
				TargetPort: intstr.FromInt(8302),
			},
			{
				Name:       "server",
				Port:       8300,
				TargetPort: intstr.FromInt(8300),
			},
			{
				Name:       "dns-tcp",
				Protocol:   "TCP",
				Port:       8600,
				TargetPort: intstr.Parse("dns-tcp"),
			},
			{
				Name:       "dns-udp",
				Protocol:   "UDP",
				Port:       8600,
				TargetPort: intstr.Parse("dns-udp"),
			},
		},
		ClusterIP: "None",
		Selector:  headlessDNSServiceLabels,
		Labels:    headlessDNSServiceLabels,

		// This must be set in addition to publishNotReadyAddresses due
		// to an open issue where it may not work:
		// https://github.com/kubernetes/kubernetes/issues/58662
		Annotations: map[string]string{
			"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
		},
	}

	uiServiceLabels := buildLabels(map[string]string{"component": "ui"})
	uiServiceSelectorLabels := buildLabels(map[string]string{"component": "server"})

	uiServiceOpts := objects.ServiceOptions{
		Name: opts.Name + "-ui",
		Ports: []v1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromInt(80),
			},
		},
		Selector: uiServiceSelectorLabels,
		Labels:   uiServiceLabels,
	}

	daemonSetFSGroup := int64(1000)
	daemonSetRunAsGroup := int64(1000)
	daemonSetRunAsUser := int64(100)
	daemonSetRunAsNonRoot := true
	daemonSetTerminationGracePeriod := int64(10)

	daemonSetSelectorLabels := buildLabels(map[string]string{"component": "client", "hasDNS": "true"})

	daemonContainerOpts := objects.ContainerOptions{
		Name:  opts.Name,
		Image: opts.Image,
		Command: []string{
			"/bin/sh",
			"-ec",
			`CONSUL_FULLNAME="consul"
exec /bin/consul agent \
-node="${NODE}" \
-advertise="${ADVERTISE_IP}" \
-bind=0.0.0.0 \
-client=0.0.0.0 \
-node-meta=pod-name:${HOSTNAME} \
-node-meta=host-ip:${HOST_IP} \
-hcl='leave_on_terminate = true' \
-disable-host-node-id=false \
-hcl='ports { grpc = 8502 }' \
-config-dir=/consul/config \
-datacenter=dc1 \
-data-dir=/consul/data \
-retry-join="${CONSUL_FULLNAME}-server-0.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc:8301" \
-retry-join="${CONSUL_FULLNAME}-server-1.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc:8301" \
-retry-join="${CONSUL_FULLNAME}-server-2.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc:8301" \
-domain=consul`,
		},
		Env: []v1.EnvVar{
			{
				Name: "ADVERTISE_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
			{
				Name: "NAMESPACE",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name: "NODE",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "spec.nodeName",
					},
				},
			},
			{
				Name: "HOST_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			},
			{
				Name: "ADVERTISE_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
		Ports: []v1.ContainerPort{
			{
				Name:          "http",
				HostPort:      8500,
				ContainerPort: 8500,
			},
			{
				Name:          "grpc",
				HostPort:      8502,
				ContainerPort: 8502,
			},
			{
				Name:          "serflan-tcp",
				ContainerPort: 8301,
				Protocol:      "TCP",
			},
			{
				Name:          "serflan-udp",
				ContainerPort: 8301,
				Protocol:      "UDP",
			},
			{
				Name:          "dns-tcp",
				ContainerPort: 8600,
				Protocol:      "TCP",
			},
			{
				Name:          "dns-udp",
				ContainerPort: 8600,
				Protocol:      "UDP",
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "data",
				MountPath: "/consul/data",
			},
			{
				Name:      "config",
				MountPath: "/consul/config",
			},
		},
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"cpu":    resource.MustParse("100m"),
				"memory": resource.MustParse("100Mi"),
			},
			Requests: v1.ResourceList{
				"cpu":    resource.MustParse("100m"),
				"memory": resource.MustParse("100Mi"),
			},
		},
		ReadinessProbe: v1.Probe{
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{
						"/bin/sh",
						"-ec",
						`curl http://127.0.0.1:8500/v1/status/leader 2>/dev/null | grep -E '".+"'`,
					},
				},
			},
		},

		// this object adds security context, restartPolicy, imagePullPolicy - keep an eye later in case it needs
		// to be excluded specifically
	}

	daemonPodTemplateOpts := objects.PodTemplateOptions{
		Name: opts.Name,
		Containers: []v1.Container{
			*objects.Container(daemonContainerOpts),
		},
		ServiceAccountName: opts.Name + "-client",
		Volumes: []v1.Volume{
			{
				Name: "data",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{
						Medium:    "",
						SizeLimit: nil,
					},
				},
			},
			{
				Name: "config",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: opts.Name + "-client-config",
						},
					},
				},
			},
		},
		Annotations: map[string]string{
			"consul.hashicorp.com/connect-inject":  "false",
			"consul.hashicorp.com/config-checksum": "79a5182ad9e59fbe2c176f4664689df261624661463a84b25746128c9bfae6b5",
		},
		Labels:                        daemonSetSelectorLabels,
		TerminationGracePeriodSeconds: &daemonSetTerminationGracePeriod,
		SecurityContext: v1.PodSecurityContext{ // I don't know if I love how this is set up.
			RunAsUser:    &daemonSetRunAsUser, // Should I make an object for the set of pointers? Need to ponder
			RunAsGroup:   &daemonSetRunAsGroup,
			RunAsNonRoot: &daemonSetRunAsNonRoot,
			FSGroup:      &daemonSetFSGroup,
		},
	}

	daemonSetOpts := objects.DaemonSetOptions{
		Name:        opts.Name,
		PodTemplate: *objects.PodTemplate(daemonPodTemplateOpts),
		Selector:    daemonSetSelectorLabels,
		Labels:      buildLabels(nil),
	}

	statefulSetPodTemplateLabels := buildLabels(map[string]string{"component": "server", "hasDNS": "true"})
	statefulSetLabels := buildLabels(map[string]string{"component": "server"})
	var statefulTerminationGracePeriod int64 = 30

	statefulPodContainerOpts := objects.ContainerOptions{
		Name:  opts.Name,
		Image: opts.Image,
		Command: []string{
			"/bin/sh",
			"-ec",
			`CONSUL_FULLNAME="consul"
	exec /bin/consul agent \
	-advertise="${ADVERTISE_IP}" \
	-bind=0.0.0.0 \
	-bootstrap-expect=3 \
	-client=0.0.0.0 \
	-config-dir=/consul/config \
	-datacenter=dc1 \
	-data-dir=/consul/data \
	-domain=consul \
	-hcl="connect { enabled = true }" \
	-ui \
	-retry-join="${CONSUL_FULLNAME}-server-0.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc:8301" \
	-retry-join="${CONSUL_FULLNAME}-server-1.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc:8301" \
	-retry-join="${CONSUL_FULLNAME}-server-2.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc:8301" \
	-serf-lan-port=8301 \
	-server`,
		},
		Env: []v1.EnvVar{
			{
				Name: "ADVERTISE_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
			{
				Name: "POD_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
			{
				Name: "NAMESPACE",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
		Ports: []v1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 8500,
			},
			{
				Name:          "serflan-tcp",
				ContainerPort: 8301,
				Protocol:      "TCP",
			},
			{
				Name:          "serflan-udp",
				ContainerPort: 8301,
				Protocol:      "UDP",
			},
			{
				Name:          "serfwan",
				ContainerPort: 8302,
			},
			{
				Name:          "server",
				ContainerPort: 8300,
			},
			{
				Name:          "dns-tcp",
				ContainerPort: 8600,
				Protocol:      "TCP",
			},
			{
				Name:          "dns-udp",
				ContainerPort: 8600,
				Protocol:      "UDP",
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "data",
				MountPath: "/consul/data",
			},
			{
				Name:      "config",
				MountPath: "/consul/config",
			},
		},
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"cpu":    resource.MustParse("100m"),
				"memory": resource.MustParse("100Mi"),
			},
			Requests: v1.ResourceList{
				"cpu":    resource.MustParse("100m"),
				"memory": resource.MustParse("100Mi"),
			},
		},
		ReadinessProbe: v1.Probe{
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{
						"/bin/sh",
						"-ec",
						`curl http://127.0.0.1:8500/v1/status/leader 2>/dev/null | grep -E '".+"'`,
					},
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      5,
			PeriodSeconds:       3,
			SuccessThreshold:    1,
			FailureThreshold:    2,
		},
	}

	statefulPodTemplateOpts := objects.PodTemplateOptions{
		Name: opts.Name + "-server",
		Containers: []v1.Container{
			*objects.Container(statefulPodContainerOpts),
		},
		ServiceAccountName: opts.Name + "-server",
		Volumes: []v1.Volume{
			{
				Name: "config",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: opts.Name + "-server-config",
						},
					},
				},
			},
		},
		Annotations: map[string]string{
			"consul.hashicorp.com/connect-inject":  "false",
			"consul.hashicorp.com/config-checksum": "f4cc930774b6eaf5726d8d6648dde5334b064308bb8a67761fcab29620737264",
		},
		Labels:                        statefulSetPodTemplateLabels,
		TerminationGracePeriodSeconds: &statefulTerminationGracePeriod,
		Affinity: v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: statefulSetLabels,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
		SecurityContext: v1.PodSecurityContext{ // I don't know if I love how this is set up.
			RunAsUser:    &daemonSetRunAsUser, // Should I make an object for the set of pointers? Need to ponder
			RunAsGroup:   &daemonSetRunAsGroup,
			RunAsNonRoot: &daemonSetRunAsNonRoot,
			FSGroup:      &daemonSetFSGroup,
		}}

	statefulPVCOpts := objects.PersistentVolumeOptions{
		Name: "data",
		AccessModes: []v1.PersistentVolumeAccessMode{
			"ReadWriteOnce",
		},
		Storage: "5Gi",
	}

	statefulSetOpts := objects.StatefulSetOptions{
		Name:        "consul-server",
		Replicas:    opts.Replicas,
		PodTemplate: *objects.PodTemplate(statefulPodTemplateOpts),
		PVCs: []v1.PersistentVolumeClaim{
			*objects.PersistentVolumeClaim(statefulPVCOpts),
		},
		Selector:            statefulSetPodTemplateLabels,
		Labels:              statefulSetLabels,
		PodManagementPolicy: "Parallel",
	}

	manifest.Append(
		objects.ServiceAccount(clientServiceAccountOpts),
		objects.ServiceAccount(serverServiceAccountOpts),
		objects.PodDisruptionBudget(podDisruptionBudgetOpts),
		objects.Role(clientRoleOpts),
		objects.Role(serverRoleOpts),
		objects.ConfigMap(clientConfigMapOpts),
		objects.ConfigMap(serverConfigMapOpts),
		objects.RoleBinding(clientRoleBindingOpts),
		objects.RoleBinding(serverRoleBindingOpts),
		objects.Service(dnsServiceOpts),
		objects.Service(headlessDNSServiceOpts),
		objects.Service(uiServiceOpts),
		objects.DaemonSet(daemonSetOpts),
		objects.StatefulSet(statefulSetOpts),
	)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}

func buildLabels(labels map[string]string) map[string]string {
	baseLabels := map[string]string{
		"app": "consul",
	}
	for k, v := range labels {
		baseLabels[k] = v
	}
	return baseLabels
}
