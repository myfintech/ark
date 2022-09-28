package kube

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObservableResource provides a container for passing around resource kinds and names
type ObservableResource struct {
	Kind string
	Name string
}

// GetObservableResourceNamesByLabel returns a map of resources that have a specific label
// this is used to provide name and kind information to the RolloutStatus function
func GetObservableResourceNamesByLabel(client Client, namespace, labelKey, labelValue string) ([]ObservableResource, error) {
	ctx := context.TODO()
	resourceSlice := make([]ObservableResource, 0)
	labelSelector := fmt.Sprintf("%s=%s", labelKey, labelValue)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	clientSet, err := client.Factory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	deploymentList, err := clientSet.AppsV1().Deployments(namespace).List(ctx, listOptions)
	if err != nil {
		return resourceSlice, err
	}
	for _, item := range deploymentList.Items {
		resourceSlice = append(resourceSlice, ObservableResource{
			Kind: "Deployment",
			Name: item.Name,
		})
	}

	daemonSetList, err := clientSet.AppsV1().DaemonSets(namespace).List(ctx, listOptions)
	if err != nil {
		return resourceSlice, err
	}
	for _, item := range daemonSetList.Items {
		resourceSlice = append(resourceSlice, ObservableResource{
			Kind: "DaemonSet",
			Name: item.Name,
		})
	}

	statefulSetList, err := clientSet.AppsV1().StatefulSets(namespace).List(ctx, listOptions)
	if err != nil {
		return resourceSlice, err
	}
	for _, item := range statefulSetList.Items {
		resourceSlice = append(resourceSlice, ObservableResource{
			Kind: "StatefulSet",
			Name: item.Name,
		})
	}
	return resourceSlice, nil
}

// GetPodsByLabel returns a slice of pod objects that match a label
// FIXME: this function should use the k8s namespace on the client
func GetPodsByLabel(client Client, namespace, labelKey, labelValue string) ([]v1.Pod, error) {
	labelSelector := fmt.Sprintf("%s=%s", labelKey, labelValue)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	clientSet, err := client.Factory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	podList, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}
