package objects

// NamespaceOptions represents fields that can be passed in to create a namespace
type NamespaceOptions struct {
	Name   string
	Labels map[string]string
}

// Namespace returns a pointer to a namespace object
