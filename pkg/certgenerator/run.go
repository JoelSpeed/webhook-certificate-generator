package certgenerator

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/joelspeed/webhook-certificate-generator/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// Run exectues the main logic of the cert generator
func Run(c *Config) error {
	client, err := utils.NewClientset(c.InCluster, c.Kubeconfig)
	if err != nil {
		return fmt.Errorf("couldn't create clientset: %v", err)
	}

	// Fetch the secret from Kubernetes.
	secret, err := utils.GetSecret(client, c.Namespace, c.SecretName)
	if err != nil {
		return fmt.Errorf("failed to fetch secret: %v", err)
	}

	// Create Kubernetes CSR
	csrName, err := createCerificateSigningRequest(client, secret, c.Namespace, c.ServiceName, c.SecretName)
	if err != nil {
		return fmt.Errorf("couldn't create certificate signing request: %v", err)
	}

	glog.Infof("Created CSR %s", csrName)

	// Approve CSR if AutoApprove enabled
	if c.AutoApprove {
		glog.Infof("Approving CSR %s", csrName)
		_, err = utils.ApproveCSR(client, csrName)
		if err != nil {
			return fmt.Errorf("couldn't approve CSR: %v", err)
		}
		glog.Infof("CSR Approved %s", csrName)
	} else {
		// Wait for CSR to be approved
		glog.Infof("Waiting for CSR approval...")
		err = waitForCSRApproval(client, csrName)
		if err != nil {
			return fmt.Errorf("error waiting for CSR approval: %v", err)
		}
		glog.Infof("CSR Approved %s", csrName)
	}

	// Wait for Certificate to be generated
	glog.Infof("Waiting for Certificate...")
	err = waitForCertificate(client, csrName)
	if err != nil {
		return fmt.Errorf("error waiting for Certificate: %v", err)
	}

	certificate, err := utils.GetCertificate(client, csrName)
	if err != nil {
		return fmt.Errorf("failed to get certificate: %v", err)
	}
	glog.Infof("Fetched Certificate")

	// Update secret
	secret.Data["cert.pem"] = certificate
	_, err = utils.CreateSecret(client, secret)
	if err != nil {
		return fmt.Errorf("couldn't create secret: %v", err)
	}
	glog.Infof("Created secret %s", secret.Name)

	if c.PatchMutating != "" {
		glog.Infof("Patching Mutating Webhook Configuration %s", c.PatchMutating)
		err = patchMutating(client, c.PatchMutating, c.Namespace, c.ServiceName)
		if err != nil {
			return fmt.Errorf("failed to patch mutating webhook configuration: %v", err)
		}
	}

	if c.PatchValidating != "" {
		glog.Infof("Patching PatchValidating Webhook Configuration %s", c.PatchValidating)
		err = patchValidating(client, c.PatchValidating, c.Namespace, c.ServiceName)
		if err != nil {
			return fmt.Errorf("failed to patch validating webhook configuration: %v", err)
		}
	}

	glog.Infof("Run complete")
	return nil
}

// Config holds required parameters
type Config struct {
	InCluster  bool   // Running inside Kubernetes cluster
	Kubeconfig string // Kubeconfig file to read from

	Namespace   string // Namespace for service and secret
	ServiceName string // Service name to generate certificate for
	SecretName  string // Secret name to store generated cert in

	AutoApprove bool // Auto Approve CSR

	PatchMutating   string // Name of MutatingWebhookConfiguration to patch CABundle
	PatchValidating string // Name of ValidatingWebhookConfiguration to patch CABundle
}

// waitForCSRApproval waits until the CSR has been approved
func waitForCSRApproval(client *kubernetes.Clientset, csrName string) error {
	return wait.PollImmediate(time.Second*10, time.Minute*10, func() (bool, error) {
		csr, err := utils.GetCSR(client, csrName)
		if err != nil {
			return false, fmt.Errorf("couldn't get CSR: %v", err)
		}
		if utils.IsCSRApproved(csr) {
			return true, nil
		}
		glog.Infof("Waiting for CSR approval...")
		return false, nil
	})
}

// waitForCertificate waits for the certificate to be ready
func waitForCertificate(client *kubernetes.Clientset, csrName string) error {
	return wait.PollImmediate(time.Second*10, time.Minute*10, func() (bool, error) {
		csr, err := utils.GetCSR(client, csrName)
		if err != nil {
			return false, fmt.Errorf("couldn't get CSR: %v", err)
		}
		if !utils.IsCSRApproved(csr) {
			return false, fmt.Errorf("cannot fetch certificate, CSR not approved")
		}

		if len(csr.Status.Certificate) > 0 {
			return true, nil
		}

		glog.Infof("Waiting for Certificate...")
		return false, nil
	})
}
