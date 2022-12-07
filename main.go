package main

import (
	"flag"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/client"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/config"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/manager"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/notify"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/resource"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

func main() {
	cmd := &cobra.Command{
		Short:   "nora-resource-detector finds unmanaged resources in kubernetes cluster.",
		Use:     "nora-resource-detector",
		Example: "  nora-resource-detector --config config.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			process()
		},
	}

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	util.AddFlags(cmd)
	if err := cmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}

func process() {
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
			ns := uns.GetNamespace()
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
