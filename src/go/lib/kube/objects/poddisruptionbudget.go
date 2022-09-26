package objects

import (
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PodDisruptionBudgetOptions represents fields that can be passed in to create a pod disruption budget
type PodDisruptionBudgetOptions struct {
	Name           string
	MaxUnavailable int32
	Labels         map[string]string
	MatchLabels    map[string]string
	Selector       metav1.LabelSelector
}

// PodDisruptionBudget returns a pointer to a pod disruption budget
func PodDisruptionBudget(opts PodDisruptionBudgetOptions) *policyv1beta1.PodDisruptionBudget {
	selector := metav1.LabelSelector{}

	if &opts.Selector != nil && opts.MatchLabels != nil { // a struct doesn't really have a 'zero' value, so the internet said to use a pointer nil check instead - is that the best way to do this?
		if opts.Selector.MatchLabels == nil {
			opts.Selector.MatchLabels = opts.MatchLabels
		} else {
			for k, v := range opts.Selector.MatchLabels {
				opts.Selector.MatchLabels[k] = v
			}
		}
		selector = opts.Selector
	} else if &opts.Selector != nil {
		selector = opts.Selector
	} else if opts.MatchLabels != nil {
		selector.MatchLabels = opts.MatchLabels
	}

	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector:       &selector,
			MaxUnavailable: &intstr.IntOrString{IntVal: opts.MaxUnavailable},
		},
	}
}
