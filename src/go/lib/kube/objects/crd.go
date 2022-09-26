package objects

import (
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomResourceDefinitionOptions represents fields that can be passed in to create a custom resource definition
type CustomResourceDefinitionOptions struct {
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Kind        string
	Group       string
	Plural      string
	Scope       v1.ResourceScope
}

// CustomResourceDefinition returns a pointer to a custom resource definition object
func CustomResourceDefinition(opts CustomResourceDefinitionOptions) *v1.CustomResourceDefinition {
	if opts.Scope == "" {
		opts.Scope = "Namespaced"
	}
	return &v1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: opts.Group,
			Names: v1.CustomResourceDefinitionNames{
				Plural: opts.Plural,
				Kind:   opts.Kind,
			},
			Scope: opts.Scope,
		},
		Status: v1.CustomResourceDefinitionStatus{
			Conditions: nil,
			AcceptedNames: v1.CustomResourceDefinitionNames{
				Plural: opts.Plural,
				Kind:   opts.Kind,
			},
		},
	}
}
