package main

import (
	"flag"

	"github.com/showcase-gig-platform/nora-resource-detector/pkg/client"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/config"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/manager"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/notify"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/resource"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
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

	detector, err := manager.InitDetector(cfg.ResourceManagers)
	if err != nil {
		klog.Fatal(err)
	}

	var unmanagedResources []util.GroupResourceName
	for _, target := range cfg.TargetResources {
		gvr, err := kc.SearchResource(target)
		if err != nil {
			klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}

		rs, err := resource.ListUnstructuredResources(kc.Client, gvr)
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
			ns := resource.MustNestedString(uns, "metadata", "namespace")
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
