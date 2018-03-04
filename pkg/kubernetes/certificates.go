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
