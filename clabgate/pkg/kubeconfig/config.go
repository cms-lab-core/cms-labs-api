// Package kubeconfig resolves Kubernetes client configuration for Clabgate.
package kubeconfig

import (
	"fmt"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type inClusterConfigLoader func() (*rest.Config, error)

// Load prefers the pod ServiceAccount configuration and falls back to the
// standard kubeconfig loading rules for local development and CI.
func Load() (*rest.Config, error) {
	return load(rest.InClusterConfig, os.Getenv(clientcmd.RecommendedConfigPathEnvVar))
}

func load(inCluster inClusterConfigLoader, kubeconfigPath string) (*rest.Config, error) {
	config, inClusterErr := inCluster()
	if inClusterErr == nil {
		return config, nil
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		loadingRules.ExplicitPath = kubeconfigPath
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load Kubernetes config (in-cluster: %v; kubeconfig: %w)", inClusterErr, err)
	}
	return config, nil
}
