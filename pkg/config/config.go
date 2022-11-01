package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

type Config struct {
	TargetResources           []string `yaml:"targetResources"`
	ManagedResourceDefinition `yaml:"managedResourceDefinition"`
	Notifier                  NotifierConfig `yaml:"notifier"`
}

type ManagedResourceDefinition struct {
	ArgoCD            ArgoCDConfig            `yaml:"argocd"`
	EksAddon          EksAddonconfig          `yaml:"eksAddon"`
	HasOwnerReference HasOwnerReferenceConfig `yaml:"hasOwnerReference"`
	StaticConfigs     []StaticConfig          `yaml:"staticConfigs"`
}

type ArgoCDConfig struct {
	InstanceLabelKey string `yaml:"instanceLabelKey"`
}

type EksAddonconfig struct{}

type HasOwnerReferenceConfig struct{}

type StaticConfig struct {
	Resource string   `yaml:"resource"`
	Names    []string `yaml:"names"`
}

type NotifierConfig struct {
	Stdout StdoutConfig `yaml:"stdout"`
}

type StdoutConfig struct{}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err != nil {
			return cfg, err
		} else {
			path = strings.ReplaceAll(path, "~", home)
		}
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to load config : %s", err)
	}

	err = yaml.UnmarshalStrict(buf, &cfg)

	// ArgoCDのinstanceLabelKeyが指定されてない場合デフォルトを設定
	// https://argo-cd.readthedocs.io/en/stable/faq/#why-is-my-app-out-of-sync-even-after-syncing
	if cfg.ArgoCD.InstanceLabelKey == "" {
		cfg.ArgoCD.InstanceLabelKey = "app.kubernetes.io/instance"
	}

	return cfg, err
}
