package utils

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	// Import authentication plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientset creates a kubernetes clientset.
// Will load config from Pod environment if running in cluster.
func NewClientset(inCluster bool, kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := createConfig(inCluster, kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return kubernetes.NewForConfig(config)
}

func createConfig(inCluster bool, kubeconfig string) (*rest.Config, error) {
	if inCluster && kubeconfig == "" {
		return rest.InClusterConfig()
	}
	clientConfigLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConfigLoadingRules.ExplicitPath = kubeconfig
	apiConfig, _ := clientConfigLoadingRules.Load()
	clientConfig := clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{})
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	return config, nil
}
