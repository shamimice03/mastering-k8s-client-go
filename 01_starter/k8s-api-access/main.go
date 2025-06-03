package main

import (
	"context"
	"fmt"
	"path/filepath"

	// k8s.io/api - Kubernetes resource definitions
	// Contains all the Kubernetes API objects like Pod, Service, Deployment, etc.

	// k8s.io/apimachinery - Common building blocks
	// Provides meta types and utilities used across Kubernetes APIs
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// k8s.io/client-go - Client library for Kubernetes API
	// Main client library for interacting with Kubernetes clusters
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// getExternalClusterConfig loads kubeconfig from ~/.kube/config
// This function is used to connect to external Kubernetes clusters
// by reading the standard kubectl configuration file
func getExternalClusterConfig() (*rest.Config, error) {
	var kubeconfig string

	// Determine the path to the kubeconfig file
	// Typically located at ~/.kube/config (standard kubectl location)
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Build config from kubeconfig file
	// This parses the YAML kubeconfig and creates a rest.Config object
	// The first parameter is for master URL override (empty means use kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
	}

	return config, nil
}

func main() {
	// Get external cluster configuration
	// This establishes connection parameters to the Kubernetes API server
	config, err := getExternalClusterConfig()
	if err != nil {
		panic(fmt.Errorf("failed to get external cluster config: %v", err))
	}

	// Create clientset to interact with Kubernetes API
	// Clientset provides access to all Kubernetes API groups and versions
	// It's the main interface for performing CRUD operations on K8s resources
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to create clientset: %v", err))
	}

	// Display successful connection information
	fmt.Printf("Connected to external cluster: %s\n", config.Host)

	// List all pods in the "default" namespace
	// context.TODO() is used when context is needed but not available
	// In production, use context.WithTimeout() or context.WithCancel()
	podList, _ := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})

	// Iterate through the list of pods and display their names
	// podList.Items contains an array of Pod objects
	for _, pod := range podList.Items {
		fmt.Printf("Pod Name: %s\n", pod.Name)
	}
}
