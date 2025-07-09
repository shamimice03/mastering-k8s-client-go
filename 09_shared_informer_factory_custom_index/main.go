package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
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
	// This factory will manage all our informers efficiently
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	// Setup Pod informer with custom indexes for efficient querying
	setupInformersWithCustomIndex(factory)

	// Create stop channel to control informer lifecycle
	stopCh := make(chan struct{})

	// Start all registered informers - they begin watching API server
	factory.Start(stopCh)

	// Wait for all informer caches to sync with current cluster state
	fmt.Println("Waiting for cache sync...")
	factory.WaitForCacheSync(stopCh)
	fmt.Println("Cache sync completed!")

	// Perform custom indexer queries on cached data
	queryWithCustomIndexers(factory)

	// Close the channel to signal all informers to stop and exit program
	close(stopCh)
}

// setupInformersWithCustomIndex creates Pod informer and adds custom indexes
func setupInformersWithCustomIndex(factory informers.SharedInformerFactory) {
	// Get Pod informer from factory
	podInformer := factory.Core().V1().Pods()

	// Add custom indexers to enable O(1) lookups by specific fields
	podInformer.Informer().AddIndexers(
		cache.Indexers{
			// Index pods by the node they're running on
			"node": func(obj interface{}) ([]string, error) {
				pod := obj.(*corev1.Pod)
				return []string{pod.Spec.NodeName}, nil
			},
			// Index pods by their current phase (Running, Pending, etc.)
			"phase": func(obj interface{}) ([]string, error) {
				pod := obj.(*corev1.Pod)
				return []string{string(pod.Status.Phase)}, nil
			},
			// Additional custom indexes can be added here
		})
}

// queryWithCustomIndexers demonstrates how to use custom indexes for efficient queries
func queryWithCustomIndexers(factory informers.SharedInformerFactory) {
	// Get the indexer from Pod informer to perform custom queries
	indexer := factory.Core().V1().Pods().Informer().GetIndexer()

	// Query 1: Get all unique node names that have pods
	allNodes := indexer.ListIndexFuncValues("node")
	fmt.Printf("Available nodes: %v\n", allNodes)

	// Query 2: Get all pods on the first available node
	if len(allNodes) > 0 {
		nodeName := allNodes[0]
		// Use custom "node" index for O(1) lookup
		podsOnNode, err := indexer.ByIndex("node", nodeName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Pods on node '%s': %d\n", nodeName, len(podsOnNode))

		// Extract and display pod names and namespaces
		for _, obj := range podsOnNode {
			pod := obj.(*corev1.Pod)
			fmt.Printf("  - %s (namespace: %s)\n", pod.Name, pod.Namespace)
		}
	} else {
		fmt.Println("No nodes found")
	}

	// Query 3: Get all pods in "Running" phase using custom index
	runningPods, err := indexer.ByIndex("phase", "Running")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Running pods: %d\n", len(runningPods))
}
