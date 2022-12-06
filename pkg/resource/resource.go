package resource

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

func ListUnstructuredResources(i dynamic.Interface, gvr schema.GroupVersionResource) (map[string]unstructured.Unstructured, error) {
	var result = map[string]unstructured.Unstructured{}
	uns, err := i.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to list resources: %s", err.Error())
		return result, err
	}
	for _, resource := range uns.Items {
		name := resource.GetName()
		result[name] = resource
	}
	return result, nil
}
