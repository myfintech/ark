package pkg

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Options struct {
	statefulapp.Options
}

func NewKafkaManifest(opts Options) (string, error) {
	var name = "kafka"

	buff := new(bytes.Buffer)
	manifest := new(objects.Manifest)
	volumeClaims := persistentVolumeClaims()

	zookeeperContainer := objects.Container(objects.ContainerOptions{
		Name:  "zookeeper",
		Image: "confluentinc/cp-zookeeper:5.4.6",
		Env: []v1.EnvVar{
			{Name: "ZOOKEEPER_CLIENT_PORT", Value: "2181"},
			{Name: "ZOOKEEPER_TICK_TIME", Value: "2000"},
		},
		VolumeMounts:    getVolumeMountByName(opts)("zookeeper"), // volumeMounts,
		ImagePullPolicy: "IfNotPresent",
	})

	brokerContainer := objects.Container(objects.ContainerOptions{
		Name:  "broker",
		Image: "confluentinc/cp-server:5.4.6",
		Env: []v1.EnvVar{
			{Name: "KAFKA_BROKER_ID", Value: "1"},
			{Name: "KAFKA_ZOOKEEPER_CONNECT", Value: "localhost:2181"},
			{Name: "KAFKA_INTER_BROKER_LISTENER_NAME", Value: "INTERNAL"},
			{Name: "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", Value: "INTERNAL:PLAINTEXT,EXTERNAL:PLAINTEXT"},
			{Name: "KAFKA_ADVERTISED_LISTENERS", Value: "INTERNAL://:9092,EXTERNAL://broker:29092"},
			{Name: "KAFKA_METRIC_REPORTERS", Value: "io.confluent.metrics.reporter.ConfluentMetricsReporter"},
			{Name: "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "CONFLUENT_BALANCER_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "CONFLUENT_DURABILITY_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "CONFLUENT_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "CONFLUENT_TIER_METADATA_REPLICATION_FACTOR", Value: "1"},
			{Name: "KAFKA_DEST_CONFLUENT_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "TRANSACTION_STATE_LOG_REPLICATION_FACTOR", Value: "1"},
			{Name: "OFFSETS_TOPIC_REPLICATION_FACTOR", Value: "1"},
			{Name: "KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS", Value: "0"},
			{Name: "CONFLUENT_METRICS_REPORTER_BOOTSTRAP_SERVERS", Value: "localhost:9092"},
			{Name: "CONFLUENT_METRICS_REPORTER_ZOOKEEPER_CONNECT", Value: "localhost:2181"},
			{Name: "CONFLUENT_METRICS_REPORTER_TOPIC_REPLICAS", Value: "1"},
			{Name: "CONFLUENT_METRICS_ENABLE", Value: "true"},
			{Name: "CONFLUENT_SUPPORT_CUSTOMER_ID", Value: "anonymous"},
		},
		VolumeMounts:    getVolumeMountByName(opts)("kafka"), // volumeMounts,
		ImagePullPolicy: "IfNotPresent",
	})

	readinessProbe := objects.Probe(objects.ProbeOptions{
		Handler: objects.Handler{
			TCPSocket: &objects.TCPSocketAction{
				Port: 9092,
			},
		},
		TimeoutSeconds:      10,
		PeriodSeconds:       5,
		InitialDelaySeconds: 40,
	})
	brokerContainer.ReadinessProbe = &readinessProbe

	livenessProbe := objects.Probe(objects.ProbeOptions{
		Handler: objects.Handler{
			Exec: &objects.ExecAction{
				Command: []string{
					"sh",
					"-c",
					"kafka-broker-api-versions --bootstrap-server localhost:9092",
				},
			},
		},
		TimeoutSeconds:      10,
		PeriodSeconds:       5,
		InitialDelaySeconds: 70,
	})
	brokerContainer.LivenessProbe = &livenessProbe

	controlCenterContainer := objects.Container(objects.ContainerOptions{
		Name:  "control-center",
		Image: "confluentinc/cp-enterprise-control-center:5.4.6",
		Env: []v1.EnvVar{
			{Name: "CONTROL_CENTER_BOOTSTRAP_SERVERS", Value: "localhost:9092"},
			{Name: "CONTROL_CENTER_ZOOKEEPER_CONNECT", Value: "localhost:2181"},
			{Name: "CONTROL_CENTER_REPLICATION_FACTOR", Value: "1"},
			{Name: "CONTROL_CENTER_INTERNAL_TOPICS_PARTITIONS", Value: "1"},
			{Name: "CONTROL_CENTER_MONITORING_INTERCEPTOR_TOPIC_PARTITIONS", Value: "1"},
			{Name: "CONFLUENT_METRICS_TOPIC_REPLICATION", Value: "1"},
			{Name: "PORT", Value: "9021"},
		},
		VolumeMounts:    getVolumeMountByName(opts)("confluent-control-center"), // volumeMounts,
		ImagePullPolicy: "IfNotPresent",
	})

	podTemp := objects.PodTemplate(objects.PodTemplateOptions{
		Name: "kafka",
		Containers: []v1.Container{
			*zookeeperContainer,
			*brokerContainer,
			*controlCenterContainer,
		},
		RestartPolicy: "Always",
		Labels:        map[string]string{"app": name},
	})

	statefulSet := objects.StatefulSet(objects.StatefulSetOptions{
		Name:          "kafka",
		Replicas:      1,
		PodTemplate:   *podTemp,
		RestartPolicy: "Always",
		PVCs:          *volumeClaims,
	})

	brokerService := objects.Service(objects.ServiceOptions{
		Name: "broker",
		Ports: []v1.ServicePort{
			{Name: "broker",
				Protocol: "TCP",
				Port:     29092,
				TargetPort: intstr.IntOrString{
					IntVal: 29092,
				}},
		},
		Type:     "ClusterIP",
		Selector: map[string]string{"app": name},
		Labels:   map[string]string{"app": name},
	})

	controlUIService := objects.Service(objects.ServiceOptions{
		Name: "control-center-ui",
		Ports: []v1.ServicePort{
			{Name: "control-center",
				Protocol: "TCP",
				Port:     9021,
				TargetPort: intstr.IntOrString{
					IntVal: 9021,
				}},
		},
		Type:     "ClusterIP",
		Selector: map[string]string{"app": name},
		Labels:   map[string]string{"app": name},
	})

	manifest.Append(statefulSet, brokerService, controlUIService)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil

}

func persistentVolumeClaims() *[]v1.PersistentVolumeClaim {
	pvcKafkaOpts := objects.PersistentVolumeOptions{
		Name:    "kafka",
		Storage: "5Gi",
	}
	pvcZooKeeperOpts := objects.PersistentVolumeOptions{
		Name:    "zookeeper",
		Storage: "5Gi",
	}
	pvcControlCenterOpts := objects.PersistentVolumeOptions{
		Name:    "confluent-control-center",
		Storage: "5Gi",
	}
	pvcs := objects.BuildVolumeClaimsTemplate([]objects.PersistentVolumeOptions{pvcKafkaOpts, pvcZooKeeperOpts, pvcControlCenterOpts})
	return pvcs
}

func computePath(opts Options, name string) string {
	path := filepath.Join("/var", "lib", name)
	if opts.DataDir != "" {
		return filepath.Join(path, opts.DataDir)
	}
	return path
}

func computePathData(opts Options, name string) string {
	return filepath.Join(computePath(opts, name), "data")
}

func computePathLog(opts Options, name string) string {
	return filepath.Join(computePath(opts, name), "log")
}

func volumeMount(opts Options) []v1.VolumeMount {
	volumeMounts := make([]v1.VolumeMount, 0)
	for _, pvc := range *persistentVolumeClaims() {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      pvc.Name,
			MountPath: computePathData(opts, pvc.Name),
			SubPath:   "data",
		}, v1.VolumeMount{
			Name:      pvc.Name,
			MountPath: computePathLog(opts, pvc.Name),
			SubPath:   "log",
		})
	}
	return volumeMounts
}

func getVolumeMountByName(opts Options) func(name string) []v1.VolumeMount {
	return func(name string) []v1.VolumeMount {
		var result []v1.VolumeMount
		for _, mnt := range volumeMount(opts) {
			if strings.Contains(mnt.Name, name) {
				result = append(result, mnt)
			}
		}
		return result
	}
}
