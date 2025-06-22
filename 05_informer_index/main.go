package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// createClientset creates and returns a Kubernetes clientset
func createClientset() *kubernetes.Clientset {
	// Get home directory for kubeconfig path
	home, err := os.UserHomeDir()
	if err != nil {
		klog.Fatalf("Failed to get home directory: %v", err)
	}
	// Parse kubeconfig flag
	kubeconfig := flag.String("kubeconfig", filepath.Join(home, "/.kube/config"), "location of kubeconfig file")
	flag.Parse()
	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build config: %v", err)
	}
	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create clientset: %v", err)
	}
	return clientset
}

// Custom indexer function: extracts node name from pod for indexing
func podNodeIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*corev1.Pod)
	return []string{pod.Spec.NodeName}, nil
}

// createPodInformer creates and returns a SharedIndexInformer for pods
func createPodInformer(clientset *kubernetes.Clientset) cache.SharedIndexInformer {
	// Create SharedIndexInformer with ListWatch functions
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			// List function - gets initial state of pods
			ListWithContextFunc: func(ctx context.Context, options v1.ListOptions) (runtime.Object, error) {
				return clientset.CoreV1().Pods("").List(context.TODO(), options)
			},
			// Watch function - creates streaming connection for pod changes
			WatchFuncWithContext: func(ctx context.Context, options v1.ListOptions) (watch.Interface, error) {
				return clientset.CoreV1().Pods("").Watch(context.TODO(), options)
			},
		},
		&corev1.Pod{},  // Object type to watch
		time.Second*30, // Resync period
		cache.Indexers{
			"namespace": cache.MetaNamespaceIndexFunc, // Built-in namespace indexer
			"node":      podNodeIndexFunc,             // Custom node indexer
		}, // Custom indexers
	)
	return informer
}

func main() {
	clientset := createClientset()
	// Create pod informer
	podInformer := createPodInformer(clientset)
	// Create stop channel for graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)
	// Start informers in background
	go podInformer.Run(stopCh)
	// Wait for caches to sync with initial data
	fmt.Println("Waiting for caches to sync...")
	if !cache.WaitForCacheSync(stopCh, podInformer.HasSynced) {
		klog.Fatal("Failed to sync caches")
	}

	// Inefficient: O(n) search through all pods without indexing
	// Get stores and list all cached objects
	podStore := podInformer.GetStore()
	allPods := podStore.List()
	fmt.Printf("Found %d pods\n", len(allPods))
	// Print pod details
	for _, obj := range allPods {
		pod := obj.(*corev1.Pod)
		fmt.Printf("Pod: %s/%s \nNode: %s\n\n", pod.Namespace, pod.Name, pod.Spec.NodeName)
	}

	// Efficient: O(1) lookup using namespace index
	fmt.Println("\n=== With Namespace Index ===")
	indexer := podInformer.GetIndexer()
	defaultPods, err := indexer.ByIndex("namespace", "default")
	if err != nil {
		fmt.Printf("Error getting indexed values: %v\n", err)
	}
	fmt.Printf("Pods in default namespace: %d\n", len(defaultPods))

	// Extract pod details from namespace index results
	fmt.Println("Pod details from namespace index:")
	for _, obj := range defaultPods {
		pod := obj.(*corev1.Pod)
		fmt.Printf("  Name: %s, Namespace: %s, Node: %s\n",
			pod.Name, pod.Namespace, pod.Spec.NodeName)
	}

	// Efficient: O(1) lookup using custom node index
	fmt.Println("\n=== With Node Index ===")
	nodeName := "k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited"
	podsOnNode, err := indexer.ByIndex("node", nodeName)
	if err != nil {
		fmt.Printf("Error getting indexed values: %v\n", err)
	}

	// Extract pod details from node index results
	fmt.Printf("Pods on node %s: %d\n", nodeName, len(podsOnNode))
	for _, obj := range podsOnNode {
		pod := obj.(*corev1.Pod)
		fmt.Printf("  Name: %s, Namespace: %s\n", pod.Name, pod.Namespace)
	}
}
