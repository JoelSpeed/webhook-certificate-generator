package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"github.com/joelspeed/webhook-certificate-generator/pkg/certgenerator"
	"github.com/spf13/cobra"
)

var (
	config *certgenerator.Config
)

func main() {
	cmd := &cobra.Command{
		Use:   "webhook-certificate-generator",
		Short: "Generate Certificates for Kubernetes Webhooks",
		Long:  `Generates Certificate for Kubernetes Webhhok admission controllers.`,
		RunE: func(c *cobra.Command, args []string) error {
			return certgenerator.Run(config)
		},
	}

	config = &certgenerator.Config{}

	cmd.Flags().BoolVar(&config.InCluster, "in-cluster", true, "Running inside a Kubernetes Cluster")
	cmd.Flags().StringVarP(&config.Kubeconfig, "kubeconfig", "k", "", "Kubeconfig file to use")

	cmd.Flags().StringVarP(&config.Namespace, "namespace", "n", "", "Service Namespace")
	cmd.Flags().StringVarP(&config.ServiceName, "service-name", "s", "", "Service to generate certificate for")
	cmd.Flags().StringVarP(&config.SecretName, "secret-name", "o", "", "Secret name to put certificates in")

	cmd.Flags().BoolVarP(&config.AutoApprove, "auto-approve-csr", "a", false, "Auto approve CSR once created")

	cmd.Flags().StringVar(&config.PatchMutating, "patch-mutating", "", "Name of MutatingWebhookConfiguration to patch CABundle into")
	cmd.Flags().StringVar(&config.PatchValidating, "patch-validating", "", "Name of ValidatingWebhookConfiguration to patch CABundle into")
	cmd.Flags().StringVar(&config.PatchApiService, "patch-api-service", "", "Name of APIService to patch CABundle into")

	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.PersistentFlags().Set("logtostderr", "True")
	flag.CommandLine.Parse([]string{})

	if err := cmd.Execute(); err != nil {
		glog.Flush()
		os.Exit(1)
	}
}
