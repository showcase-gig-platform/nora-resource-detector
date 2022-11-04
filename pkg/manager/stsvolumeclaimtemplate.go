package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
	"strings"

	"github.com/showcase-gig-platform/nora-resource-detector/pkg/resource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type StsVolumeClaimTemplateConfig struct{}

type StsVolumeClaimTemplateDetector struct {
	prefixes []string
}

func NewStsVolumeClaimTemplateDetector(i dynamic.Interface) StsVolumeClaimTemplateDetector {
	prefixes, err := stsVolumeClaimTemplatePrefixes(i)
	if err != nil {
		klog.Errorf("failed to get statefulset persistent volume claim prefixes: %s", err.Error())
	}
	return StsVolumeClaimTemplateDetector{
		prefixes,
	}
}

func (sd StsVolumeClaimTemplateDetector) Execute(uns unstructured.Unstructured) bool {
	name := resource.MustNestedString(uns, "metadata", "name")
	for _, prefix := range sd.prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func stsVolumeClaimTemplatePrefixes(i dynamic.Interface) ([]string, error) {
	var result []string
	uns, err := i.Resource(schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "statefulsets",
	}).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return result, fmt.Errorf("failed to get statefulset: %s", err.Error())
	}
	for _, item := range uns.Items {
		stsName := resource.MustNestedString(item, "metadata", "name")
		vcts := resource.MustNestedSlice(item, "spec", "volumeClaimTemplates")
		for _, vct := range vcts {
			pvc, err := fetchPvcTemplate(vct)
			if err != nil {
				klog.Errorf("failed to fetch persistentVolumeClaimTemplate: %s", err.Error())
			}
			result = append(result, fmt.Sprintf("%s-%s-", pvc.ObjectMeta.Name, stsName))
		}
	}
	return result, nil
}

func fetchPvcTemplate(i interface{}) (v1.PersistentVolumeClaimTemplate, error) {
	pvct := v1.PersistentVolumeClaimTemplate{}
	data, err := json.Marshal(i)
	if err != nil {
		return pvct, err
	}
	err = json.Unmarshal(data, &pvct)
	if err != nil {
		return pvct, err
	}
	return pvct, nil
}
