package objects

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeOptions represents fields that can be passed in to create a persistent volume/persistent volume claim
type PersistentVolumeOptions struct {
	Name        string
	AccessModes []corev1.PersistentVolumeAccessMode
	Storage     string
	Annotations map[string]string
}

// PersistentVolumeClaim returns a pointer to a persistent volume claim
func PersistentVolumeClaim(opts PersistentVolumeOptions) *corev1.PersistentVolumeClaim {
	defaultLabels := map[string]string{"app": opts.Name}
	if opts.AccessModes == nil {
		opts.AccessModes = make([]corev1.PersistentVolumeAccessMode, 0)
		opts.AccessModes = append(opts.AccessModes, "ReadWriteOnce")
	}
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      defaultLabels,
			Annotations: opts.Annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: opts.AccessModes,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": resource.MustParse(opts.Storage),
				},
			},
		},
	}
}

// PersistentVolume returns a persistent volume

// BuildVolumeClaimsTemplate returns a pointer to a slice of PVCs - for use with stateful sets
func BuildVolumeClaimsTemplate(opts []PersistentVolumeOptions) *[]corev1.PersistentVolumeClaim {
	claimTemplateSlice := make([]corev1.PersistentVolumeClaim, 0)
	for _, pvc := range opts {
		claim := PersistentVolumeClaim(pvc)
		claimTemplateSlice = append(claimTemplateSlice, *claim)
	}
	return &claimTemplateSlice
}
