package main

import (
	"os"

	"github.com/joelspeed/webhook-certificate-generator/pkg/certgenerator"
	"github.com/joelspeed/webhook-certificate-generator/pkg/kubernetes"
	"github.com/spf13/cobra"
)

var (
	inCluster  bool
	kubeconfig string

	namespace   string
	serviceName string
	secretName  string
)

func main() {
	cmd := &cobra.Command{
		Use:   "webhook-certificate-generator",
		Short: "Generate Certificates for Kubernetes Webhooks",
		Long:  `Generates Certificate for Kubernetes Webhhok admission controllers.`,
		RunE: func(c *cobra.Command, args []string) error {
			client, err := kubernetes.NewClientset(inCluster, kubeconfig)
			if err != nil {
				return err
			}
			return certgenerator.CreateCerificateSigningRequest(client, namespace, serviceName, secretName)
		},
	}

	cmd.Flags().BoolVar(&inCluster, "in-cluster", true, "Running inside a Kubernetes Cluster")
	cmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Kubeconfig file to use")

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Service Namespace")
	cmd.Flags().StringVarP(&serviceName, "serivce-name", "s", "", "Service to generate certificate for")
	cmd.Flags().StringVarP(&secretName, "secret-name", "o", "", "Secret name to put certificates in")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
