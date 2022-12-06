package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
	"strings"

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

// stsのvolumeClaimTemplateから作成されたpvcにはownerReferenceが付かず（sts削除時にgcしないので）、
// どこのリソースから派生したpvcなのか厳密に判定することはできない
// 一応、`template名`-`sts名`-`0からの連番` （`sts名`-`0からの連番` = pod名） という命名規則はあるので、
// 「`template名`-`sts名`-」 がnameのprefixになっているpvcは管理されたpvcと判定する
// なので、手動作成した野良pvcがこの命名規則に当てはまると野良リソースとして検出できない

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
	name := uns.GetName()
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
		stsName := item.GetName()
		vcts, _, e := unstructured.NestedSlice(item.Object, "spec", "volumeClaimTemplates")
		if e != nil {
			klog.Errorf("failed to get spec.volumeClaimTemplates from statefulset: %s", err.Error())
		}
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
