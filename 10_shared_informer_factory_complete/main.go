package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	// Parse kubeconfig flag
	kubeconfig := flag.String("kubeconfig", filepath.Join(home, "/.kube/config"), "location of kubeconfig file")
	flag.Parse()

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	return clientset
}

func main() {
	// Create Kubernetes clientset
	clientset := createClientSet()

	// Create SharedInformerFactory with 30-second resync period
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	// Setup custom indexers
	setupCustomIndexers(factory)

	// Setup event handlers
	setupPodMonitor(factory)

	// Start and wait for sync
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// Query using listers and custom indexes
	queryBylisters(factory)
	queryByCustomIndexes(factory)

	// Block until program termination
	<-stopCh
}

// setupCustomIndexers adds custom indexing functions to the pod informer
func setupCustomIndexers(factory informers.SharedInformerFactory) {
	// Get pod informer
	podInformer := factory.Core().V1().Pods()

	// Add custom indexer that indexes pods by node name
	podInformer.Informer().AddIndexers(cache.Indexers{
		"node": func(obj interface{}) ([]string, error) {
			pod := obj.(*corev1.Pod)
			return []string{pod.Spec.NodeName}, nil
		},
	})
}

// setupPodMonitor configures event handlers for pod events
func setupPodMonitor(factory informers.SharedInformerFactory) {
	// Get pod informer
	podInformer := factory.Core().V1().Pods()

	// Add event handler for pod additions
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("Pod added: %s\n", pod.Name)
		},
	})
}

// queryBylisters demonstrates querying using listers
func queryBylisters(factory informers.SharedInformerFactory) {
	// Get pod lister
	podLister := factory.Core().V1().Pods().Lister()

	// Query by namespace
	defaultPods, _ := podLister.Pods("default").List(labels.Everything())
	fmt.Printf("Default namespace pods: %d\n", len(defaultPods))

	// Query by labels
	labelSelector, _ := labels.Parse("app=nginx")
	nginxPods, _ := podLister.List(labelSelector)
	fmt.Printf("Nginx pods: %d\n", len(nginxPods))
}

// queryByCustomIndexes demonstrates querying using custom indexes
func queryByCustomIndexes(factory informers.SharedInformerFactory) {
	// Get indexer from pod informer
	indexer := factory.Core().V1().Pods().Informer().GetIndexer()

	// Query by custom node index
	allNodes := indexer.ListIndexFuncValues("node")
	fmt.Printf("Nodes: %v\n", allNodes)

	if len(allNodes) > 0 {
		// Get pods on first node using custom index
		podsOnNode, _ := indexer.ByIndex("node", allNodes[0])
		fmt.Printf("Pods on %s: %d\n", allNodes[0], len(podsOnNode))
	}
}
