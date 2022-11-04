package notify

import (
	"encoding/json"
	"fmt"

	"github.com/showcase-gig-platform/nora-resource-detector/pkg/util"
	"k8s.io/klog/v2"
)

type StdoutConfig struct{}

type StdoutNotifier struct{}

func NewStdoutNotifier() StdoutNotifier {
	return StdoutNotifier{}
}

func (s StdoutNotifier) notify(results []util.GroupResourceName) {
	for _, result := range results {
		jsonbytes, err := json.Marshal(result)
		if err != nil {
			klog.Errorf("failed to marshal result json: %s", err.Error())
		}
		fmt.Println(string(jsonbytes))
	}
}
