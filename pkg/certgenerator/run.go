package certgenerator

import (
	"fmt"

	"github.com/joelspeed/webhook-certificate-generator/pkg/kubernetes"
)

// Run exectues the main logic of the cert generator
func Run(c *Config) error {
	client, err := kubernetes.NewClientset(c.InCluster, c.Kubeconfig)
	if err != nil {
		return fmt.Errorf("couldn't create clientset: %v", err)
	}
	// Create Kubernetes CSR
	csrName, err := CreateCerificateSigningRequest(client, c.Namespace, c.ServiceName, c.SecretName)
	if err != nil {
		return fmt.Errorf("couldn't create certificate signing request: %v", err)
	}

	// Approve CSR if AutoApprove enabled
	if c.AutoApprove {
		_, err = kubernetes.ApproveCSR(client, csrName)
		if err != nil {
			return fmt.Errorf("couldn't approve CSR: %v", err)
		}
	}
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
}
