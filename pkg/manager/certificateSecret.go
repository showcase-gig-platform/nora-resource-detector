package manager

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

type CertificateSecretConfig struct{}

type CertificateSecretDetector struct {
	certificates []string
}

func NewCertificateSecretDetector(i dynamic.Interface) CertificateSecretDetector {
	cs, err := certificateNames(i)
	if err != nil {
		klog.Errorf("failed to get certificate name list: %s", err.Error())
	}
	klog.Infoln(cs)
	return CertificateSecretDetector{
		certificates: cs,
	}
}

func (cd CertificateSecretDetector) Execute(uns unstructured.Unstructured) bool {
	ans := uns.GetAnnotations()
	target, ok := ans["cert-manager.io/certificate-name"]
	if !ok {
		return false
	}
	for _, certificate := range cd.certificates {
		if certificate == target {
			return true
		}
	}
	return false
}

func certificateNames(i dynamic.Interface) ([]string, error) {
	var result []string
	uns, err := i.Resource(schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1", // TODO: versionは指定しなくていいようにしたい
		Resource: "certificates",
	}).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return result, fmt.Errorf("failed to get certificates : %s", err.Error())
	}
	for _, item := range uns.Items {
		result = append(result, item.GetName())
	}
	return result, nil
}
