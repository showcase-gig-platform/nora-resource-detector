package util

import (
	"github.com/spf13/cobra"
)

var (
	ConfigPath         string
	ApiserverUrl       string
	Kubeconfig         string
	KubeContext        string
	UseInclusterConfig bool
)

func AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", "~/.nora/config.yaml", "Path to config file.")
	cmd.PersistentFlags().StringVarP(&ApiserverUrl, "apiserver-url", "u", "", "URL for kubernetes api server.")
	cmd.PersistentFlags().StringVarP(&Kubeconfig, "kubeconfig", "k", "", "Path to kubeconfig file.")
	cmd.PersistentFlags().StringVarP(&KubeContext, "context", "", "", "Kubeconfig context name to use.")
	cmd.PersistentFlags().BoolVarP(&UseInclusterConfig, "in-cluster", "i", false, "Set true if used in kubernetes cluster.")
}
