package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kav1beta1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1beta1"
	aggregatorclient "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

// GetAPIServiceConfiguration gets the names api service configuration
// from kubernetes
func GetAPIServiceConfiguration(aggrclient *aggregatorclient.Clientset, name string) (*kav1beta1.APIService, error) {
	getOpts := metav1.GetOptions{}
	return aggrclient.ApiregistrationV1beta1().APIServices().Get(name, getOpts)
}

// UpdateAPIServiceCongiguration updates the api service configuration
// given
func UpdateAPIServiceConfiguration(aggrclient *aggregatorclient.Clientset, svc *kav1beta1.APIService) (*kav1beta1.APIService, error) {
	return aggrclient.ApiregistrationV1beta1().APIServices().Update(svc)
}
