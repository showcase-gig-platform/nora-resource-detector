package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/config"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/notify"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	configPath   string
	apiserverUrl string
	kubeconfig   string
)

const (
	managedFieldsEKSManagerName = "eks"
)

type GroupResourceName struct {
	Group    string
	Resource string
	Name     string
}

type fetchedStaticConfig struct {
	schema.GroupVersionResource
	names []string
}

func init() {
	flag.StringVar(&configPath, "config", "~/.nora/config.yaml", "Path to config file.")
	flag.StringVar(&apiserverUrl, "apiserver-url", "", "URL for kubernetes api server.")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file.")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof("loaded config: %#v", cfg)

	clientConfig, err := clientcmd.BuildConfigFromFlags(apiserverUrl, kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build kubeconfig: %s", err.Error())
	}

	rm, err := restMapper(clientConfig)
	if err != nil {
		klog.Fatalf("Failed to get resource discovery mapper: %s", err.Error())
	}

	client, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		klog.Fatalf("kubernetes.NewForConfig failed: %v", err)
	}

	argos, err := argoApplications(client)
	if err != nil {
		klog.Errorf("failed to get argocd applications: %s", err.Error())
	}

	var unmanagedResources []GroupResourceName
	for _, resource := range cfg.TargetResources {
		gvr, err := searchResource(rm, resource)
		if err != nil {
			klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}
		rs, err := listUnstructuredResources(client, gvr)
		if err != nil {
			klog.Errorf("failed to list resources : %s", err.Error())
			continue
		}

		for name, uns := range rs {
			if managedByArgoCD(uns, argos, cfg.ArgoCD.InstanceLabelKey) {
				delete(rs, name)
			}

			if hasOwnerReferences(uns) {
				delete(rs, name)
			}

			if eksAddon(uns) {
				delete(rs, name)
			}

			if staticConfigResource(name, gvr, fetchStaticConfig(cfg.StaticConfigs, rm)) {
				delete(rs, name)
			}
		}

		for name, _ := range rs {
			unmanagedResources = append(unmanagedResources, GroupResourceName{
				Group:    gvr.Group,
				Resource: gvr.Resource,
				Name:     name,
			})
		}
	}
	klog.Infoln(unmanagedResources)
	n := notify.NewStdoutNotifier()
	notify.Notify(n)
}

func restMapper(c *rest.Config) (apimeta.RESTMapper, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(c)
	if err != nil {
		return nil, err
	}
	gr, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	return mapper, nil
}

func searchResource(rm apimeta.RESTMapper, resource string) (schema.GroupVersionResource, error) {
	gvr, err := rm.ResourceFor(schema.GroupVersionResource{
		Group:    "",
		Version:  "",
		Resource: resource,
	})

	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return gvr, nil
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
		name, ok, err := unstructured.NestedString(item.Object, "metadata", "name")
		if ok {
			result = append(result, name)
		} else {
			if err != nil {
				klog.Errorf("failed to fetch argocd application name: %s", err.Error())
			} else {
				klog.Errorln("failed to fetch argocd application name with no error.")
			}
		}
	}
	return result, nil
}

func managedByArgoCD(uns unstructured.Unstructured, applications []string, labelKey string) bool {
	md, ok, _ := unstructured.NestedMap(uns.Object, "metadata", "labels")
	if ok {
		target := md[labelKey]
		s, aok := target.(string)
		if aok {
			for _, application := range applications {
				if s == application {
					return true
				}
			}
		}
	}
	return false
}

func hasOwnerReferences(uns unstructured.Unstructured) bool {
	_, ok, _ := unstructured.NestedSlice(uns.Object, "metadata", "ownerReferences")
	if ok {
		return true
	} else {
		return false
	}
}

func eksAddon(uns unstructured.Unstructured) bool {
	mf, ok, _ := unstructured.NestedSlice(uns.Object, "metadata", "managedFields")
	if ok {
		for _, ifc := range mf {
			m, ok := ifc.(map[string]interface{})
			if ok {
				if m["manager"] == managedFieldsEKSManagerName {
					return true
				}
			}
		}
	}
	return false
}

func fetchStaticConfig(confs []config.StaticConfig, rm apimeta.RESTMapper) []fetchedStaticConfig {
	var result []fetchedStaticConfig
	for _, conf := range confs {
		gvr, err := searchResource(rm, conf.Resource)
		if err != nil {
			klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}
		result = append(result, fetchedStaticConfig{
			GroupVersionResource: gvr,
			names:                conf.Names,
		})
	}
	return result
}

func staticConfigResource(name string, gvr schema.GroupVersionResource, confs []fetchedStaticConfig) bool {
	for _, conf := range confs {
		if conf.GroupVersionResource == gvr {
			for _, s := range conf.names {
				if name == s {
					return true
				}
			}
		}
	}
	return false
}
