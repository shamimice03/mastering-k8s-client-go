package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	clientset := createClientSet()

	_, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to connect to cluster: %v", err)
	}
	fmt.Println("Successfully connected to cluster")
	// Single factory for all informers
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	// Setup informers (this registers them with the factory)
	setupInformers(factory)
	// Start all informers at once
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	// defer factory.Shutdown()

	// Wait for cache sync
	fmt.Println("Waiting for cache sync...")
	factory.WaitForCacheSync(stopCh)
	fmt.Println("Cache sync completed!")

	// Now demonstrate listers
	demonstrateListers(factory)

	fmt.Println("Lister demonstration completed!")

	// Close the channel to stop informers and exit
	close(stopCh)
}

func setupInformers(factory informers.SharedInformerFactory) {
	// Create the informers so they get registered with the factory
	// Must call .Informer() to actually register them for syncing
	factory.Core().V1().Pods().Informer()
	factory.Apps().V1().Deployments().Informer()
}

func demonstrateListers(factory informers.SharedInformerFactory) {
	// Get listers
	podLister := factory.Core().V1().Pods().Lister()
	deploymentLister := factory.Apps().V1().Deployments().Lister()

	// Query cached data efficiently

	// 1. Get ALL pods (across all namespaces)
	allPods, err := podLister.List(labels.Everything())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Total pods (all namespaces): %d\n", len(allPods))

	// Show pods by namespace
	namespaceCount := make(map[string]int)
	for _, pod := range allPods {
		namespaceCount[pod.Namespace]++
	}
	fmt.Println("Pods per namespace:")
	for ns, count := range namespaceCount {
		fmt.Printf("  %s: %d pods\n", ns, count)
	}

	// 2. Get pods in default namespace specifically
	defaultPods, err := podLister.Pods("default").List(labels.Everything())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Pods in default namespace: %d\n", len(defaultPods))

	// 3. Get specific pod by name (if any pods exist)
	if len(allPods) > 0 {
		firstPod := allPods[0]
		pod, err := podLister.Pods(firstPod.Namespace).Get(firstPod.Name)
		if err != nil {
			fmt.Printf("Pod not found: %v\n", err)
		} else {
			fmt.Printf("Found pod: %s in namespace: %s\n", pod.Name, pod.Namespace)
		}
	}

	// 4. Filter by labels (try different labels if nginx doesn't exist)
	labelSelector, _ := labels.Parse("app=nginx")
	nginxPods, err := podLister.List(labelSelector)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Nginx pods: %d\n", len(nginxPods))

	// Try other common labels
	if len(nginxPods) == 0 {
		// Try k8s-app label (common in system pods)
		labelSelector, _ = labels.Parse("k8s-app")
		systemPods, err := podLister.List(labelSelector)
		if err == nil {
			fmt.Printf("Pods with k8s-app label: %d\n", len(systemPods))
		}
	}

	// 5. Query ALL deployments
	allDeployments, err := deploymentLister.List(labels.Everything())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Total deployments (all namespaces): %d\n", len(allDeployments))

	// Show deployments by namespace
	if len(allDeployments) > 0 {
		fmt.Println("Deployments:")
		for _, deployment := range allDeployments {
			fmt.Printf("  %s/%s (replicas: %d)\n",
				deployment.Namespace, deployment.Name, *deployment.Spec.Replicas)
		}
	}
}
