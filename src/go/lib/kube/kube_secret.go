package kube

import (
	"context"

	"k8s.io/client-go/rest"

	"k8s.io/kubectl/pkg/scheme"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateSecret either creates a secret if one does not exist OR it updates a secret that does exist, overwriting its data
func CreateOrUpdateSecret(client Client, namespace, secretName string, data, labels map[string]string) (result *corev1.Secret, err error) {
	result = &corev1.Secret{}
	restClient, err := client.Factory.RESTClient()
	if err != nil {
		return result, err
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels:    labels,
		},
		StringData: data, // the StringData field is a write-only field that allows for raw strings rather than having to base64 encode the data ahead of writing it
	}

	if _, err = GetSecret(restClient, namespace, secretName); err != nil {
		return CreateSecret(restClient, namespace, secret)
	}

	return UpdateSecret(restClient, namespace, secret)
}

// GetSecret checks for the existence of a secret
func GetSecret(client *rest.RESTClient, namespace, secretName string) (result *corev1.Secret, err error) {
	result = &corev1.Secret{}
	err = client.Get().
		Namespace(namespace).
		Resource("secrets").
		Name(secretName).
		VersionedParams(&metav1.UpdateOptions{}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	return
}

// CreateSecret creates a secret
func CreateSecret(client *rest.RESTClient, namespace string, secret *corev1.Secret) (result *corev1.Secret, err error) {
	result = &corev1.Secret{}
	err = client.Post().
		Namespace(namespace).
		Resource("secrets").
		VersionedParams(&metav1.UpdateOptions{}, scheme.ParameterCodec).
		Body(secret).
		Do(context.TODO()).
		Into(result)
	return
}

// UpdateSecret updates a secret that already exists
func UpdateSecret(client *rest.RESTClient, namespace string, secret *corev1.Secret) (result *corev1.Secret, err error) {
	result = &corev1.Secret{}
	err = client.Put().
		Namespace(namespace).
		Resource("secrets").
		Name(secret.Name).
		VersionedParams(&metav1.UpdateOptions{}, scheme.ParameterCodec).
		Body(secret).
		Do(context.TODO()).
		Into(result)
	return
}

// DeleteSecret removes a secret
func DeleteSecret(client *rest.RESTClient, namespace, secretName string) error {
	return client.Delete().
		Namespace(namespace).
		Resource("secrets").
		Name(secretName).
		Do(context.TODO()).
		Error()
}
