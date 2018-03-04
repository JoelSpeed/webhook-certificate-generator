package certgenerator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	wcgkubernetes "github.com/joelspeed/webhook-certificate-generator/pkg/kubernetes"
	certsv1beta1 "k8s.io/api/certificates/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateCerificateSigningRequest creates a Kubernetes CSR for the given service
func CreateCerificateSigningRequest(client *kubernetes.Clientset, namespace string, serviceName string, secretName string) (string, error) {
	secret, err := wcgkubernetes.GetSecret(client, namespace, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to load secret: %v", err)
	}
	csrPem, err := createCSRPem(secret, namespace, serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to create CSR Pem: %v", err)
	}

	csr := &certsv1beta1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", serviceName, namespace),
		},
		Spec: certsv1beta1.CertificateSigningRequestSpec{
			Request: csrPem,
			Usages: []certsv1beta1.KeyUsage{
				certsv1beta1.UsageDigitalSignature,
				certsv1beta1.UsageKeyEncipherment,
				certsv1beta1.UsageServerAuth,
			},
		},
	}
	_, err = wcgkubernetes.CreateCSR(client, csr)
	if err != nil {
		return "", fmt.Errorf("failed to create CSR: %v", err)
	}
	return csr.Name, nil
}

func createCSRPem(secret *v1.Secret, namespace string, serviceName string) ([]byte, error) {
	privateKey, err := getPrivateKey(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %v", err)
	}

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		},
		DNSNames: []string{
			serviceName,
			fmt.Sprintf("%s.%s", serviceName, namespace),
			fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		},
	}

	caExt, err := createCAExtension()
	if err != nil {
		return nil, fmt.Errorf("failed to create CA Extension")
	}

	template.Extensions = []pkix.Extension{caExt}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr}), nil
}

func getPrivateKey(secret *v1.Secret) (*rsa.PrivateKey, error) {
	if keyPem64, ok := secret.Data["key.pem"]; ok {
		keyPem, err := base64.StdEncoding.DecodeString(string(keyPem64))
		if err != nil {
			return nil, fmt.Errorf("failed to decode secret: %v", err)
		}
		pemBlock, _ := pem.Decode(keyPem)
		if pemBlock == nil {
			return nil, fmt.Errorf("faileed to decode private key pem: %v", err)
		}
		return x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	}
	return rsa.GenerateKey(rand.Reader, 2048)

}

// BasicConstraints CSR information RFC 5280, 4.2.1.9
type BasicConstraints struct {
	IsCA       bool `asn1:"optional"`
	MaxPathLen int  `asn1:"optional,default:-1"`
}

func createCAExtension() (pkix.Extension, error) {
	val, err := asn1.Marshal(BasicConstraints{false, 0})
	if err != nil {
		return pkix.Extension{}, fmt.Errorf("failed to marshal basic constraints: %v", err)
	}

	return pkix.Extension{
		Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
		Value:    val,
		Critical: true,
	}, nil
}
