package objects

import (
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// Manifest represents a container for runtime objects that can be encoded into a valid Kubernetes manifest
type Manifest struct {
	CoreList v1.List
}

// Append adds a runtime object to a Manifest
func (m *Manifest) Append(objects ...runtime.Object) {
	for _, object := range objects {
		m.CoreList.Items = append(m.CoreList.Items, runtime.RawExtension{
			Object: object,
		})
	}
}

// Serialize encodes items in a manifest and writes them to a designated writer
func (m *Manifest) Serialize(stream io.Writer) error {
	encoder := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	for i, item := range m.CoreList.Items {
		if err := encoder.Encode(item.Object, stream); err != nil {
			return err
		}
		if i != len(m.CoreList.Items)-1 {
			if _, err := fmt.Fprint(stream, "---\n"); err != nil {
				return err
			}
		}
	}

	return nil
}
