package kubernetes

import (
	arv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetMutatingWebhookConfiguration gets the names mutating webhook configuration
// from kubernetes
func GetMutatingWebhookConfiguration(client *kubernetes.Clientset, name string) (*arv1beta1.MutatingWebhookConfiguration, error) {
	getOpts := metav1.GetOptions{}
	return client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(name, getOpts)
}

// UpdateMutatingWebhookConfiguration updates the mutating webhook configuration
// given
func UpdateMutatingWebhookConfiguration(client *kubernetes.Clientset, mwc *arv1beta1.MutatingWebhookConfiguration) (*arv1beta1.MutatingWebhookConfiguration, error) {
	return client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(mwc)
}

// GetValidatingWebhookConfiguration gets the names mutating webhook configuration
// from kubernetes
func GetValidatingWebhookConfiguration(client *kubernetes.Clientset, name string) (*arv1beta1.ValidatingWebhookConfiguration, error) {
	getOpts := metav1.GetOptions{}
	return client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(name, getOpts)
}

// UpdateValidatingWebhookConfiguration updates the mutating webhook configuration
// given
func UpdateValidatingWebhookConfiguration(client *kubernetes.Clientset, mwc *arv1beta1.ValidatingWebhookConfiguration) (*arv1beta1.ValidatingWebhookConfiguration, error) {
	return client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(mwc)
}
