package notify

import (
	"errors"
	"fmt"
	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
	"github.com/slack-go/slack"
	"k8s.io/klog/v2"
	"os"
)

type SlackConfig struct {
	Token     string `yaml:"token"`
	Channel   string `yaml:"channel"`
	IconEmoji string `yaml:"icon_emoji"`
	IconURL   string `yaml:"icon_url"`
	Username  string `yaml:"username"`
}

type SlackNotifier struct {
	client *slack.Client
	SlackConfig
}

func NewSlackNotifier(cfg SlackConfig) (SlackNotifier, error) {
	val, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		val = cfg.Token
	}
	if val == "" {
		return SlackNotifier{}, errors.New("please set slack token in env `SLACK_TOKEN` or config file `notifiers.slack.token`")
	}
	client := slack.New(val)
	return SlackNotifier{
		client,
		cfg,
	}, nil
}

func (s SlackNotifier) notify(results []util.GroupResourceName) {
	var options []slack.MsgOption
	attachment := slack.Attachment{}
	var fields []slack.AttachmentField
	for _, result := range results {
		fields = append(fields, slack.AttachmentField{
			Title: "",
			Value: fmt.Sprintf("Resource: %v\nNamespace: %v\nName: %v", result.Resource, result.Namespace, result.Name),
			Short: true,
		})
	}
	attachment.Fields = fields
	attachment.Title = "Nora resource detected"
	attachment.Color = "danger"

	options = append(options, slack.MsgOptionAttachments(attachment))
	if s.IconEmoji != "" {
		options = append(options, slack.MsgOptionIconEmoji(s.IconEmoji))
	}
	if s.IconURL != "" {
		options = append(options, slack.MsgOptionIconURL(s.IconURL))
	}
	if s.Username != "" {
		options = append(options, slack.MsgOptionUsername(s.Username))
	}

	_, _, err := s.client.PostMessage(s.Channel, options...)
	if err != nil {
		klog.Errorf("failed to post slack: %s", err.Error())
	}
}
