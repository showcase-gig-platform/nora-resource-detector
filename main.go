package main

import (
	"context"
	"flag"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/client"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/config"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/manager"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/notify"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	cfg, err := config.LoadConfig(util.ConfigPath)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof("loaded config: %#v", cfg)

	kc, err := client.NewKubeClient()
	if err != nil {
		klog.Fatal(err)
	}

	detector, err := manager.NewDetector(cfg.ResourceManagers)
	if err != nil {
		klog.Fatal(err)
	}

	var unmanagedResources []util.GroupResourceName
	for _, resource := range cfg.TargetResources {
		gvr, err := kc.SearchResource(resource)
		if err != nil {
			klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}

		rs, err := listUnstructuredResources(kc.Client, gvr)
		if err != nil {
			klog.Errorf("failed to list resources : %s", err.Error())
			continue
		}

		for name, uns := range rs {
			if detector.Execute(uns) {
				delete(rs, name)
			}
		}

		for name, uns := range rs {
			ns, ok, _ := unstructured.NestedString(uns.Object, "metadata", "namespace")
			if !ok {
				klog.Errorf("unstructured resource does not have `metadata.namespace` : %v", uns.Object)
			}
			unmanagedResources = append(unmanagedResources, util.GroupResourceName{
				Group:     gvr.Group,
				Resource:  gvr.Resource,
				Namespace: ns,
				Name:      name,
			})
		}
	}
	if len(unmanagedResources) == 0 {
		klog.Info("there is no nora resource!")
		return
	}

	notifier := notify.NewNotifier(cfg.Notifiers)
	notifier.Execute(unmanagedResources)
}

func listUnstructuredResources(i dynamic.Interface, gvr schema.GroupVersionResource) (map[string]unstructured.Unstructured, error) {
	var result = map[string]unstructured.Unstructured{}
	uns, err := i.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to list resources: %s", err.Error())
		return result, err
	}
	for _, resource := range uns.Items {
		name, ok, _ := unstructured.NestedString(resource.Object, "metadata", "name")
		if !ok {
			klog.Errorf("unstructured resource does not have `metadata.name` : %v", resource.Object)
		}
		result[name] = resource
	}
	return result, nil
}
