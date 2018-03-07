# Webhook Certificate Generator
Uses the Kubernetes CSR Api to create a Secret containing a private key and
signed certificate.
This tool is intended to be deployed alongside a Mutating/Validating
Webhook Admission Controller.
Since these Admission controllers need certificates with a trusted CA,
this tool signs them using the Kubernetes cluster CA and can then patch the
webhook definition appropriately.

## Usage
When creating a service `foo` in the namespace `bar`, the following will
generate a CSR and place the Certificate and Private key in a secret called
`foo-certs`:

```bash
wcg --service-name=foo --namespace=bar --secret-name=foo-certs
```

This generates a CSR with the name `foo.bar` (`servicename.namespace`) in the
K8s API and waits for it's approval.

Once the CSR is approved, `wcg` waits for the certificate to be signed and then
creates a secret `foo-certs` with the following format:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: foo-certs
data:
  cert.pem: <BASE64_ENCODED_CERT_PEM>
  key.pem: <BASE64_ENCODED_PRIVATE_KEY_PEM>
```

This secret can then be mounted into a Pod.

### CSR Approval
WCG will pause after it has generated the CSR. The CSR needs to be approved
before the secret can be generated.

To approve the CSR:
- Manually approve the CSR: `kubectl certificate approve foo.bar`
- Automatically approve the CSR: Add the flag `--auto-approve-csr` to `wcg` when
  requesting a certificate.

### Webhook Configuration Patching
In Kubernetes 1.9, Mutating and Validating Webhook Configurations have a
mandatory `caBundle` field in the service definition. This field should contain
the CA certificate chain that signs the serving certificates of your webhook
(The certificates generated by `wcg`).

Since this certificate is likely to be different per cluster, `wcg` can auto
patch this field for you.

Create the Webhook configutaion with an empty CA Bundle:
```yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: foo-webhook
webhooks:
  - name: foo.example.com
    clientConfig:
      service:
        name: foo
        namespace: bar
        path: /admissionreviews
      caBundle: ""
    ...
```

Add the `--patch-mutating` or `--patch-validating` flag when running `wcg` with
the name of the Webhook configuration (eg `--patch-mutating=foo-webhook`) and
`wcg` will patch the `caBundle` field with the cluster's CA Bundle once the
certificate has been issued.

### Flags
The following flags are configure the certificate generation process.
`namespace`, `secret-name` and `service-name` are required.
```
  -a, --auto-approve-csr                 Auto approve CSR once created
      --in-cluster                       Running inside a Kubernetes Cluster (default true)
  -k, --kubeconfig string                Kubeconfig file to use
  -n, --namespace string                 Service Namespace
      --patch-mutating string            Name of MutatingWebhookConfiguration to patch CABundle into
      --patch-validating string          Name of ValidatingWebhookConfiguration to patch CABundle into
  -o, --secret-name string               Secret name to put certificates in
  -s, --service-name string              Service to generate certificate for
```

## Kubernetes Installation
A collection of example Kubernetes Manifests are available in the
[install](install/kubernetes) folder.

These allow you to configure certificate generator as a Kubernetes Job.

The appropriate RBAC bindings have also been included.
It is recommended to run this Job in the `kube-system` namespace or another
namespace only accessible by cluster admins.
In an RBAC enabled system, the service account may be granted rights to approve
CSRs.
Make sure you understand the risks of this before you grant these privileges.

## Communication

* Found a bug? Please open an issue.
* Have a feature request. Please open an issue.
* If you want to contribute, please submit a pull request

## Contributing
Please see our [Contributing](CONTRIBUTING.md) guidelines.

## License
This project is licensed under Apache 2.0 and a copy of the license is available [here](LICENSE).
