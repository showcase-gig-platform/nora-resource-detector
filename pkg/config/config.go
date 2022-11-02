package config

import (
	"fmt"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/manager"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/notify"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

type Config struct {
	TargetResources  []string                        `yaml:"targetResources"`
	ResourceManagers []manager.ResourceManagerConfig `yaml:"resourceManagers"`
	Notifiers        []notify.NotifierConfig         `yaml:"notifiers"`
}

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

	return cfg, err
}
