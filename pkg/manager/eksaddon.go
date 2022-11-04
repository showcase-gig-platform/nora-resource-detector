package manager

import (
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type EksAddonConfig struct{}

type EksAddonDetector struct{}

const (
	managedFieldsEKSManagerName = "eks"
)

func NewEksAddonDetector() EksAddonDetector {
	return EksAddonDetector{}
}

func (ed EksAddonDetector) Execute(uns unstructured.Unstructured) bool {
	mf := resource.MustNestedSlice(uns, "metadata", "managedFields")
	for _, ifc := range mf {
		m, ok := ifc.(map[string]interface{})
		if ok {
			if m["manager"] == managedFieldsEKSManagerName {
				return true
			}
		}
	}
	return false
}
