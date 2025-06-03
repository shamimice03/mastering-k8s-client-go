package main

import (
	"context"
	"fmt"
	"path/filepath"

	// k8s.io/api - Kubernetes resource definitions
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	// k8s.io/apimachinery - Common building blocks
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// k8s.io/client-go - Client library for Kubernetes API
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// getExternalClusterConfig loads kubeconfig from ~/.kube/config
func getExternalClusterConfig() (*rest.Config, error) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
	}

	return config, nil
}

// Helper function to convert int32 to *int32
func int32Ptr(i int32) *int32 {
	return &i
}

func main() {
	// Get external cluster configuration
	config, err := getExternalClusterConfig()
	if err != nil {
		panic(fmt.Errorf("failed to get external cluster config: %v", err))
	}

	// Create clientset to interact with Kubernetes API
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to create clientset: %v", err))
	}

	fmt.Printf("Connected to external cluster: %s\n", config.Host)

	// Define a Deployment object
	deployment := &appsv1.Deployment{
		// TypeMeta - from apimachinery (API version and kind)
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		// ObjectMeta - from apimachinery (name, namespace, labels)
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "default",
		},
		// Spec - from k8s.io/api/apps/v1 (Deployment-specific configuration)
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx-app",
				},
			},
			// PodTemplateSpec - from k8s.io/api/core/v1
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx-app",
							Image: "nginx:1.21",
						},
					},
				},
			},
		},
	}

	// Create Deployment using client-go
	res, err := clientset.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		panic(fmt.Errorf("failed to create deployment: %v", err))
	}

	fmt.Printf("Successfully created deployment: %s\n", res.Name)
}
