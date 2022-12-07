package util

import (
	"flag"
)

var (
	ConfigPath         string
	ApiserverUrl       string
	Kubeconfig         string
	KubeContext        string
	UseInclusterConfig bool
)

func init() {
	flag.StringVar(&ConfigPath, "config", "~/.nora/config.yaml", "Path to config file.")
	flag.StringVar(&ApiserverUrl, "apiserver-url", "", "URL for kubernetes api server.")
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "Path to kubeconfig file.")
	flag.StringVar(&KubeContext, "context", "", "Kubeconfig context name to use.")
	flag.BoolVar(&UseInclusterConfig, "in-cluster", false, "Set true if used in kubernetes cluster.")
}
