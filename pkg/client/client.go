package client

import (
	"fmt"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"

	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type KubeClient struct {
	Mapper meta.RESTMapper
	Client dynamic.Interface
}

func NewKubeClient() (KubeClient, error) {
	var clientConfig *rest.Config
	var err error
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if util.UseInclusterConfig {
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			return KubeClient{}, fmt.Errorf("failed to load in cluster config: %s", err.Error())
		}
	}
	configFromEnv := os.Getenv("KUBECONFIG")
	if len(util.Kubeconfig) != 0 {
		kubeconfigPath = util.Kubeconfig
	} else if len(configFromEnv) != 0 {
		kubeconfigPath = configFromEnv
	}

	cor := &clientcmd.ConfigOverrides{
		ClusterInfo: api.Cluster{
			Server: util.ApiserverUrl,
		},
	}
	if len(util.KubeContext) != 0 {
		cor.CurrentContext = util.KubeContext
	}
	clientConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: kubeconfigPath,
		},
		cor,
	).ClientConfig()

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
