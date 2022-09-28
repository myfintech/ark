package k8s_echo

import corev1 "k8s.io/api/core/v1"

func podIndexByLabel(label string, pod *corev1.Pod) []string {
	var podNames []string
	_, ok := pod.GetLabels()[label]
	if ok {
		podNames = append(podNames, pod.GetName())
	}
	return podNames
}

type podInformerIndex struct{}

func (i podInformerIndex) deployedByArk(obj interface{}) ([]string, error) {
	pod := obj.(*corev1.Pod)
	return podIndexByLabel("ark.target.key", pod), nil
}

func (i podInformerIndex) withLiveSync(obj interface{}) ([]string, error) {
	pod := obj.(*corev1.Pod)
	return podIndexByLabel("ark.live.sync.enabled", pod), nil
}
