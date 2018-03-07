package utils

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetConfigMap gets the named configmap from kubernetes
func GetConfigMap(client *kubernetes.Clientset, namespace string, name string) (*v1.ConfigMap, error) {
	getOpts := metav1.GetOptions{}
	return client.CoreV1().ConfigMaps(namespace).Get(name, getOpts)
}
