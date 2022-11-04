package manager

import (
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/client"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type StaticConfig struct {
	Configs []Config `yaml:"configs"`
}

type Config struct {
	Resource  string   `yaml:"resource"`
	Namespace string   `yaml:"namespace"`
	Names     []string `yaml:"names"`
}

type StaticDetector struct {
	confs  []fetchedStaticConfig
	client client.KubeClient
}

type fetchedStaticConfig struct {
	schema.GroupVersionResource
	namespace string
	names     []string
}

func NewStaticdetector(config *StaticConfig, cli client.KubeClient) StaticDetector {
	fetchedConfig := fetchStaticConfig(config.Configs, cli)

	return StaticDetector{
		confs:  fetchedConfig,
		client: cli,
	}
}

func (sd StaticDetector) Execute(uns unstructured.Unstructured) bool {
	name := resource.MustNestedString(uns, "metadata", "name")
	ns := resource.MustNestedString(uns, "metadata", "namespace")

	kind := resource.MustNestedString(uns, "kind")

	gvr, err := sd.client.SearchResource(kind)
	if err != nil {
		klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
		return false
	}

	return staticConfigResource(ns, name, gvr, sd.confs)
}

func fetchStaticConfig(confs []Config, cli client.KubeClient) []fetchedStaticConfig {
	var result []fetchedStaticConfig
	for _, conf := range confs {
		gvr, err := cli.SearchResource(conf.Resource)
		if err != nil {
			klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}
		result = append(result, fetchedStaticConfig{
			GroupVersionResource: gvr,
			namespace:            conf.Namespace,
			names:                conf.Names,
		})
	}
	return result
}

func staticConfigResource(ns, name string, gvr schema.GroupVersionResource, confs []fetchedStaticConfig) bool {
	for _, conf := range confs {
		if conf.GroupVersionResource == gvr && ns == conf.namespace {
			for _, s := range conf.names {
				if name == s {
					return true
				}
			}
		}
	}
	return false
}
