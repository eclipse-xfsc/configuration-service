package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetConfigmap retrieves a ConfigMap from the Kubernetes cluster
func GetConfigmap(name string, namespace string) (map[string]string, error) {
	emptyConfigmap := make(map[string]string)

	config, err := rest.InClusterConfig()
	if err != nil {
		return emptyConfigmap, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return emptyConfigmap, err
	}

	configmap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return emptyConfigmap, err
	}

	return configmap.Data, nil
}
