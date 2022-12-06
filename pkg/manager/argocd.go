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

type ArgoCDConfig struct {
	InstanceLabelKey string `yaml:"instanceLabelKey"`
}

type ArgoCDDetector struct {
	applications     []string
	instanceLabelKey string
}

const (
	defaultInstanceLabelKey = "app.kubernetes.io/instance"
)

func NewArgoCDDetector(cfg *ArgoCDConfig, i dynamic.Interface) ArgoCDDetector {
	argos, err := argoApplications(i)
	if err != nil {
		klog.Errorf("failed to get argocd applications: %s", err.Error())
	}

	// ArgoCDのinstanceLabelKeyが指定されてない場合デフォルトを設定
	// https://argo-cd.readthedocs.io/en/stable/faq/#why-is-my-app-out-of-sync-even-after-syncing
	if cfg.InstanceLabelKey == "" {
		cfg.InstanceLabelKey = defaultInstanceLabelKey
	}

	return ArgoCDDetector{
		applications:     argos,
		instanceLabelKey: cfg.InstanceLabelKey,
	}
}

func (ad ArgoCDDetector) Execute(uns unstructured.Unstructured) bool {
	return managedByArgoCD(uns, ad.applications, ad.instanceLabelKey)
}

// github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned を使ってapplicationsを取得しようとしたけど
// https://github.com/argoproj/argo-cd/issues/4055 こんな感じで面倒なのでdynamicを流用
func argoApplications(i dynamic.Interface) ([]string, error) {
	var result []string
	uns, err := i.Resource(schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return result, fmt.Errorf("failed to get argocd applications: %s", err.Error())
	}
	for _, item := range uns.Items {
		result = append(result, item.GetName())
	}
	return result, nil
}

func managedByArgoCD(uns unstructured.Unstructured, applications []string, labelKey string) bool {
	md := uns.GetLabels()
	target := md[labelKey]
	for _, application := range applications {
		if target == application {
			return true
		}
	}
	return false
}
