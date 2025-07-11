package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// createClientset creates and returns a Kubernetes clientset
func createClientSet() *kubernetes.Clientset {
	// Get home directory for kubeconfig path
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	// Parse kubeconfig flag to get the path to kubeconfig file
	kubeconfig := flag.String("kubeconfig", filepath.Join(home, "/.kube/config"), "location of kubeconfig file")
	flag.Parse()
	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}
	// Create clientset from the config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}
	return clientset
}

func main() {
	// Create Kubernetes client
	clientset := createClientSet()
	// Test connection to cluster by listing namespaces
	_, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to connect to cluster: %v", err)
	}
	fmt.Println("Successfully connected to cluster")

	// Create single SharedInformerFactory with 30-second resync period
	informers.NewSharedInformerFactory(clientset, time.Second*30)

	// Create factory scoped to specific namespace
	informers.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Second*30,
		informers.WithNamespace("default"),
	)

	// Create factory filtered by label selector
	informers.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Second*30,
		informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
			lo.LabelSelector = "app=nginx"
		}),
	)

	// Create factory filtered by field selector
	informers.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Second*30,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = "status.phase=Running"
		}),
	)

	// Create factory with multiple filters combined
	informers.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Second*30,
		informers.WithNamespace("kube-system"),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = "k8s-app"
			options.FieldSelector = "status.phase=Running"
		}),
	)

	// Create factory with custom resync periods per resource type
	informers.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Second*30,
		informers.WithCustomResyncConfig(map[metav1.Object]time.Duration{
			&corev1.Pod{}:        time.Second * 10,
			&appsv1.Deployment{}: time.Second * 60,
		}),
	)
}
