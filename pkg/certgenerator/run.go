package certgenerator

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	wcgkubernetes "github.com/joelspeed/webhook-certificate-generator/pkg/kubernetes"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// Run exectues the main logic of the cert generator
func Run(c *Config) error {
	client, err := wcgkubernetes.NewClientset(c.InCluster, c.Kubeconfig)
	if err != nil {
		return fmt.Errorf("couldn't create clientset: %v", err)
	}
	// Create Kubernetes CSR
	csrName, err := CreateCerificateSigningRequest(client, c.Namespace, c.ServiceName, c.SecretName)
	if err != nil {
		return fmt.Errorf("couldn't create certificate signing request: %v", err)
	}

	glog.Infof("Created CSR %s", csrName)

	// Approve CSR if AutoApprove enabled
	if c.AutoApprove {
		glog.Infof("Approving CSR %s", csrName)
		_, err = wcgkubernetes.ApproveCSR(client, csrName)
		if err != nil {
			return fmt.Errorf("couldn't approve CSR: %v", err)
		}
		glog.Infof("CSR Approved %s", csrName)
	} else {
		// Wait for CSR to be approved
		glog.Infof("Waiting for CSR approval...")
		err = waitForCSR(client, csrName)
		if err != nil {
			return fmt.Errorf("error waiting for CSR approval: %v", err)
		}
		glog.Infof("CSR Approved %s", csrName)
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

func waitForCSR(client *kubernetes.Clientset, csrName string) error {
	return wait.Poll(time.Second*10, time.Minute*10, func() (bool, error) {
		csr, err := wcgkubernetes.GetCSR(client, csrName)
		if err != nil {
			return false, fmt.Errorf("couldn't get CSR: %v", err)
		}
		if wcgkubernetes.IsCSRApproved(csr) {
			return true, nil
		}
		glog.Infof("Waiting for CSR approval...")
		return false, nil
	})
}
