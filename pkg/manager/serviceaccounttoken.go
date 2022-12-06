package manager

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type ServiceAccountTokenConfig struct{}

type ServiceAccountTokenDetector struct{}

const (
	serviceAccountTokenType = "kubernetes.io/service-account-token"
)

func NewServiceAccountTokenDetector() ServiceAccountTokenDetector {
	return ServiceAccountTokenDetector{}
}

func (ed ServiceAccountTokenDetector) Execute(uns unstructured.Unstructured) bool {
	if uns.GetKind() != "Secret" {
		return false
	}
	st, _, err := unstructured.NestedString(uns.Object, "type")
	if err != nil {
		klog.Errorf("failed to get secret type: %s", err.Error())
		return false
	}
	if st == serviceAccountTokenType {
		return true
	}
	return false
}
