package kubernetes

import (
	"fmt"

	certsv1beta1 "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateCSR creates a Kubernetes CSR from the input given
// If the CSR already exists it will update it
func CreateCSR(client *kubernetes.Clientset, csr *certsv1beta1.CertificateSigningRequest) (*certsv1beta1.CertificateSigningRequest, error) {
	c, err := getCSRIfExists(client, csr.Name)
	if err != nil {
		return nil, fmt.Errorf("couldn't get CSR: %v", err)
	}
	if c == nil {
		return client.CertificatesV1beta1().CertificateSigningRequests().Create(csr)
	}
	return client.CertificatesV1beta1().CertificateSigningRequests().Update(csr)
}

// ApproveCSR approves the CSR with the given name
func ApproveCSR(client *kubernetes.Clientset, csrName string) (*certsv1beta1.CertificateSigningRequest, error) {
	csr, err := getCSRIfExists(client, csrName)
	if err != nil {
		return nil, fmt.Errorf("couldn't get CSR: %v", err)
	}
	if csr == nil {
		return nil, fmt.Errorf("no CSR with name %s", csrName)
	}
	if IsCSRApproved(csr) {
		return csr, nil
	}

	csr.Status.Conditions = append(csr.Status.Conditions,
		certsv1beta1.CertificateSigningRequestCondition{
			Type:           certsv1beta1.CertificateApproved,
			Reason:         "WCGApprove",
			Message:        "This CSR was approved by webhook certificate generator.",
			LastUpdateTime: metav1.Now(),
		},
	)

	return client.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(csr)
}

// GetCSR retrieves the named CSR from Kubernetes
func GetCSR(client *kubernetes.Clientset, csrName string) (*certsv1beta1.CertificateSigningRequest, error) {
	c, err := getCSRIfExists(client, csrName)
	if err != nil {
		return nil, fmt.Errorf("couldn't get CSR: %v", err)
	}
	if c == nil {
		return nil, fmt.Errorf("CSR %s not found", csrName)
	}
	return c, nil
}

// GetCertificate returns the Base64 encoded certificate from the CSR
func GetCertificate(client *kubernetes.Clientset, csrName string) ([]byte, error) {
	c, err := getCSRIfExists(client, csrName)
	if err != nil {
		return nil, fmt.Errorf("couldn't get CSR: %v", err)
	}
	if c == nil {
		return nil, fmt.Errorf("CSR %s not found", csrName)
	}
	if len(c.Status.Certificate) == 0 {
		return nil, fmt.Errorf("Certificate not yet ready")
	}
	return c.Status.Certificate, nil
}

// IsCSRApproved determines whether the CSR has been approved or not
func IsCSRApproved(csr *certsv1beta1.CertificateSigningRequest) bool {
	for _, c := range csr.Status.Conditions {
		if c.Type == certsv1beta1.CertificateApproved {
			return true
		}
	}
	return false
}

func getCSRIfExists(client *kubernetes.Clientset, name string) (*certsv1beta1.CertificateSigningRequest, error) {
	listOpts := metav1.ListOptions{}
	csrList, err := client.CertificatesV1beta1().CertificateSigningRequests().List(listOpts)
	if err != nil {
		return nil, fmt.Errorf("couldn't list CSRs: %v", err)
	}

	for _, csr := range csrList.Items {
		if csr.Name == name {
			return &csr, nil
		}
	}
	return nil, nil
}
