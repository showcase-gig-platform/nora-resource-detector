package notify

import (
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
	"k8s.io/klog/v2"
)

type NotifierConfig struct {
	Stdout *StdoutConfig `yaml:"stdout"`
	Slack  *SlackConfig  `yaml:"slack"`
}

type Notifier interface {
	notify([]util.GroupResourceName)
}

type Notifiers []Notifier

func NewNotifier(cfgs []NotifierConfig) Notifiers {
	return createNotifiers(cfgs)
}

func createNotifiers(cfgs []NotifierConfig) Notifiers {
	var notifiers Notifiers
	for _, cfg := range cfgs {
		if cfg.Stdout != nil {
			notifiers = append(notifiers, NewStdoutNotifier())
			continue
		}

		if cfg.Slack != nil {
			sn, err := NewSlackNotifier(*cfg.Slack)
			if err != nil {
				klog.Errorf("failed to create slack notifier: %s", err.Error())
				continue
			}
			notifiers = append(notifiers, sn)
			continue
		}

		klog.Error("no match NotifierConfig")
	}
	return notifiers
}

func (n Notifiers) Execute(resources []util.GroupResourceName) {
	for _, notifier := range n {
		notifier.notify(resources)
	}
}
