package certgenerator

import (
	"fmt"

	"github.com/joelspeed/webhook-certificate-generator/pkg/utils"
	arv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/client-go/kubernetes"
)

func patchMutating(client *kubernetes.Clientset, name string, namespace string, service string) error {
	caBundle, err := fetchCABundle(client)
	if err != nil {
		return fmt.Errorf("error retrieving ca bundle: %v", err)
	}

	mwc, err := utils.GetMutatingWebhookConfiguration(client, name)
	if err != nil {
		return fmt.Errorf("failed to fetch mutating webhook configuration: %v", err)
	}

	mwc.Webhooks = patchWebhooks(mwc.Webhooks, caBundle, namespace, service)
	_, err = utils.UpdateMutatingWebhookConfiguration(client, mwc)
	if err != nil {
		return fmt.Errorf("failed updating mutating webhook configuration: %v", err)
	}
	return nil
}

func patchValidating(client *kubernetes.Clientset, name string, namespace string, service string) error {
	caBundle, err := fetchCABundle(client)
	if err != nil {
		return fmt.Errorf("error retrieving ca bundle: %v", err)
	}

	vwc, err := utils.GetValidatingWebhookConfiguration(client, name)
	if err != nil {
		return fmt.Errorf("failed to fetch validating webhook configuration: %v", err)
	}

	vwc.Webhooks = patchWebhooks(vwc.Webhooks, caBundle, namespace, service)
	_, err = utils.UpdateValidatingWebhookConfiguration(client, vwc)
	if err != nil {
		return fmt.Errorf("failed updating validating webhook configuration: %v", err)
	}
	return nil
}

func fetchCABundle(client *kubernetes.Clientset) ([]byte, error) {
	cm, err := utils.GetConfigMap(client, "kube-system", "extension-apiserver-authentication")
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve auth configmap: %v", err)
	}
	if bundle, ok := cm.Data["client-ca-file"]; ok {
		return []byte(bundle), nil
	}
	return nil, fmt.Errorf("no client-ca-file in configmap")
}

func patchWebhooks(webhooks []arv1beta1.Webhook, caBundle []byte, namespace string, name string) []arv1beta1.Webhook {
	outWebhooks := []arv1beta1.Webhook{}
	for _, wh := range webhooks {
		if wh.ClientConfig.Service.Namespace == namespace &&
			wh.ClientConfig.Service.Name == name {
			wh.ClientConfig.CABundle = caBundle
		}
		outWebhooks = append(outWebhooks, wh)
	}
	return outWebhooks
}
