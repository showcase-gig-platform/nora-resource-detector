package resource

import (
	"context"
	"strings"

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
		name := MustNestedString(resource, "metadata", "name")
		result[name] = resource
	}
	return result, nil
}

func MustNestedString(uns unstructured.Unstructured, fields ...string) string {
	result, ok, err := unstructured.NestedString(uns.Object, fields...)
	if err != nil {
		klog.Errorf("failed to get nested string from unstructured object: %s", err.Error())
	}
	if !ok {
		klog.Infof("unstructured resource does not have `%v` : %v", strings.Join(fields, "."), uns.Object)
	}
	return result
}

func MustNestedSlice(uns unstructured.Unstructured, fields ...string) []interface{} {
	result, ok, err := unstructured.NestedSlice(uns.Object, fields...)
	if err != nil {
		klog.Errorf("failed to get nested slice from unstructured object: %s", err.Error())
	}
	if !ok {
		klog.Infof("unstructured resource does not have `%v` : %v", strings.Join(fields, "."), uns.Object)
	}
	return result
}

func MustNestedMap(uns unstructured.Unstructured, fields ...string) map[string]interface{} {
	result, ok, err := unstructured.NestedMap(uns.Object, fields...)
	if err != nil {
		klog.Errorf("failed to get nested map from unstructured object: %s", err.Error())
	}
	if !ok {
		klog.Infof("unstructured resource does not have `%v` : %v", strings.Join(fields, "."), uns.Object)
	}
	return result
}
