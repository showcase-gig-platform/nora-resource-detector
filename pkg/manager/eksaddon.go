package manager

import (
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
	mfs := uns.GetManagedFields()
	for _, mf := range mfs {
		if mf.Manager == managedFieldsEKSManagerName {
			return true
		}
	}
	return false
}
