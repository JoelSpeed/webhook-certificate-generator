package utils

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetSecret gets the named secret from kubernetes
// If the names secret does not exist, it will generate an empty secret
func GetSecret(client *kubernetes.Clientset, namespace string, name string) (*v1.Secret, error) {
	cm, err := getSecretIfExists(client, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve secret: %v", err)
	}
	if cm != nil {
		return cm, nil
	}
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: make(map[string][]byte),
	}, nil
}

// CreateSecret will create the secret if it does not exist
// If the secert does exist, it will update the secret
func CreateSecret(client *kubernetes.Clientset, secret *v1.Secret) (*v1.Secret, error) {
	cm, err := getSecretIfExists(client, secret.Namespace, secret.Name)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve secret: %v", err)
	}
	if cm == nil {
		// Secret doesn't currently exist
		return client.CoreV1().Secrets(secret.Namespace).Create(secret)
	}
	return client.CoreV1().Secrets(secret.Namespace).Update(secret)
}

func getSecretIfExists(client *kubernetes.Clientset, namespace string, name string) (*v1.Secret, error) {
	getOpts := metav1.GetOptions{}
	secret, err := client.CoreV1().Secrets(namespace).Get(name, getOpts)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("couldn't get secrets: %v", err)
	}
	return secret, nil
}
