package manager

import (
	"fmt"

	"github.com/showcase-gig-platform/nora-resource-detector/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type ResourceManagerConfig struct {
	ArgoCD                 *ArgoCDConfig                 `yaml:"argocd"`
	EksAddon               *EksAddonConfig               `yaml:"eksAddon"`
	OwnerReference         *OwnerReferenceConfig         `yaml:"ownerReference"`
	Static                 *StaticConfig                 `yaml:"static"`
	StsVolumeClaimTemplate *StsVolumeClaimTemplateConfig `yaml:"stsVolumeClaimTemplate"`
}

type ResourceManagerDetector interface {
	Execute(unstructured unstructured.Unstructured) bool
}

type Detector struct {
	detectors []ResourceManagerDetector
}

func InitDetector(cfg []ResourceManagerConfig) (Detector, error) {
	var kubeclient, err = client.NewKubeClient()
	if err != nil {
		return Detector{}, fmt.Errorf("failed to get kubernetes client: %s", err.Error())
	}
	var detectors []ResourceManagerDetector
	addDetectors(cfg, &detectors, kubeclient)
	return Detector{
		detectors,
	}, nil
}

func addDetectors(configs []ResourceManagerConfig, detectors *[]ResourceManagerDetector, kubeclient client.KubeClient) {
	for _, cfg := range configs {
		if cfg.ArgoCD != nil {
			*detectors = append(*detectors, NewArgoCDDetector(cfg.ArgoCD, kubeclient.Client))
			continue
		}

		if cfg.EksAddon != nil {
			*detectors = append(*detectors, NewEksAddonDetector())
			continue
		}

		if cfg.OwnerReference != nil {
			*detectors = append(*detectors, NewOwnerReferenceDetector())
			continue
		}

		if cfg.Static != nil {
			*detectors = append(*detectors, NewStaticdetector(cfg.Static, kubeclient))
			continue
		}

		if cfg.StsVolumeClaimTemplate != nil {
			*detectors = append(*detectors, NewStsVolumeClaimTemplateDetector(kubeclient.Client))
			continue
		}

		klog.Error("no match ResourceManagerConfig")
	}
}

func (d Detector) Execute(uns unstructured.Unstructured) bool {
	for _, detector := range d.detectors {
		if detector.Execute(uns) {
			return true
		}
	}
	return false
}
