package manager

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type EksAddonConfig struct{}

type EksAddonDetector struct{}

const (
	managedFieldsEKSManagerName = "eks"
)

func NewEksAddonDetector() EksAddonDetector {
	return EksAddonDetector{}
}

func (ed EksAddonDetector) Execute(uns unstructured.Unstructured) bool {
	mf, ok, _ := unstructured.NestedSlice(uns.Object, "metadata", "managedFields")
	if ok {
		for _, ifc := range mf {
			m, ok := ifc.(map[string]interface{})
			if ok {
				if m["manager"] == managedFieldsEKSManagerName {
					return true
				}
			}
		}
	}
	return false
}
