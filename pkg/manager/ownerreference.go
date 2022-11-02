package manager

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type OwnerReferenceConfig struct{}

type OwnerReferenceDetector struct{}

func NewOwnerReferenceDetector() OwnerReferenceDetector {
	return OwnerReferenceDetector{}
}

func (od OwnerReferenceDetector) Execute(uns unstructured.Unstructured) bool {
	_, ok, _ := unstructured.NestedSlice(uns.Object, "metadata", "ownerReferences")
	if ok {
		return true
	} else {
		return false
	}
}
