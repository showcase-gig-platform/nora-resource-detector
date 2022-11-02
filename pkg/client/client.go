package client

import (
	"fmt"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	Mapper meta.RESTMapper
	Client dynamic.Interface
}

func NewKubeClient() (KubeClient, error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags(util.ApiserverUrl, util.Kubeconfig)
	if err != nil {
		return KubeClient{}, fmt.Errorf("failed to build kubeconfig: %s", err.Error())
	}

	rm, err := restMapper(clientConfig)
	if err != nil {
		return KubeClient{}, fmt.Errorf("failed to get resource discovery mapper: %s", err.Error())
	}

	client, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return KubeClient{}, fmt.Errorf("kubernetes.NewForConfig failed: %v", err)
	}

	return KubeClient{
		Mapper: rm,
		Client: client,
	}, nil
}

func restMapper(c *rest.Config) (meta.RESTMapper, error) {
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

func (k KubeClient) SearchResource(resource string) (schema.GroupVersionResource, error) {
	gvr, err := k.Mapper.ResourceFor(schema.GroupVersionResource{
		Group:    "",
		Version:  "",
		Resource: resource,
	})

	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return gvr, nil
}
