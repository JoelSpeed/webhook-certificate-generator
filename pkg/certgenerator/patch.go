package certgenerator

import (
	"fmt"
	"strings"

	"github.com/joelspeed/webhook-certificate-generator/pkg/utils"
	arv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/client-go/kubernetes"
	aggregatorclient "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

func patchMutating(client *kubernetes.Clientset, name string, namespace string, service string, apiServiceGroup string) error {
	caBundle, err := fetchCABundle(client)

	if err != nil {
		return fmt.Errorf("error retrieving ca bundle: %v", err)
	}

	mwc, err := utils.GetMutatingWebhookConfiguration(client, name)
	if err != nil {
		return fmt.Errorf("failed to fetch mutating webhook configuration: %v", err)
	}

	mwc.Webhooks = patchWebhooks(mwc.Webhooks, caBundle, namespace, service, apiServiceGroup)

	_, err = utils.UpdateMutatingWebhookConfiguration(client, mwc)
	if err != nil {
		return fmt.Errorf("failed updating mutating webhook configuration: %v", err)
	}
	return nil
}

func patchValidating(client *kubernetes.Clientset, name string, namespace string, service string, apiServiceGroup string) error {
	caBundle, err := fetchCABundle(client)
	if err != nil {
		return fmt.Errorf("error retrieving ca bundle: %v", err)
	}

	vwc, err := utils.GetValidatingWebhookConfiguration(client, name)
	if err != nil {
		return fmt.Errorf("failed to fetch validating webhook configuration: %v", err)
	}

	vwc.Webhooks = patchWebhooks(vwc.Webhooks, caBundle, namespace, service, apiServiceGroup)
	_, err = utils.UpdateValidatingWebhookConfiguration(client, vwc)
	if err != nil {
		return fmt.Errorf("failed updating validating webhook configuration: %v", err)
	}
	return nil
}

func patchAPIService(kubeclient *kubernetes.Clientset, client *aggregatorclient.Clientset, name string) error {
	caBundle, err := fetchCABundle(kubeclient)
	if err != nil {
		return fmt.Errorf("error retrieving ca bundle: %v", err)
	}

	svc, err := utils.GetAPIServiceConfiguration(client, name)
	if err != nil {
		return fmt.Errorf("failed to fetch api service configuration: %v", err)
	}

	svc.Spec.InsecureSkipTLSVerify = false
	svc.Spec.CABundle = caBundle

	_, err = utils.UpdateAPIServiceConfiguration(client, svc)
	if err != nil {
		return fmt.Errorf("failed updating api service configuration: %v", err)
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

func patchWebhooks(webhooks []arv1beta1.Webhook, caBundle []byte, namespace string, name string, apiServiceGroup string) []arv1beta1.Webhook {
	outWebhooks := []arv1beta1.Webhook{}
	for _, wh := range webhooks {

		matchesService := wh.ClientConfig.Service.Namespace == namespace && wh.ClientConfig.Service.Name == name
		matchesAggregatedAPI := apiServiceGroup != "" &&
			wh.ClientConfig.Service.Namespace == "default" &&
			wh.ClientConfig.Service.Name == "kubernetes" &&
			strings.HasPrefix(*wh.ClientConfig.Service.Path, apiServiceGroup)

		if matchesService || matchesAggregatedAPI {
			wh.ClientConfig.CABundle = caBundle
		}
		outWebhooks = append(outWebhooks, wh)
	}
	return outWebhooks
}
