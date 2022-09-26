package k8s_echo

type InformerInderxer interface {
	deployedByArk(interface{}) ([]string, error)
}
