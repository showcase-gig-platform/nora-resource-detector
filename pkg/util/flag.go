package util

import (
	"flag"
)

var (
	ConfigPath   string
	ApiserverUrl string
	Kubeconfig   string
)

func init() {
	flag.StringVar(&ConfigPath, "config", "~/.nora/config.yaml", "Path to config file.")
	flag.StringVar(&ApiserverUrl, "apiserver-url", "", "URL for kubernetes api server.")
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "Path to kubeconfig file.")
}
