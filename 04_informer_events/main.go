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
		&corev1.Pod{},    // Object type to watch
		time.Second*30,   // Resync period
		cache.Indexers{}, // Custom indexers
	)
	return informer
}

func main() {
	clientset := createClientset()

	// Create ONE pod informer instance - this will be shared among multiple handlers
	podInformer := createPodInformer(clientset)

	// Create stop channel for graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

	// Start informer - creates SINGLE watch connection to API server
	go podInformer.Run(stopCh)

	// Wait for caches to sync with initial data
	fmt.Println("Waiting for caches to sync...")
	if !cache.WaitForCacheSync(stopCh, podInformer.HasSynced) {
		klog.Fatal("Failed to sync caches")
	}

	// SHARED ASPECT: First handler - multiple handlers can share the same informer
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("(+) Pod added: %s/%s\n", pod.Namespace, pod.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*corev1.Pod)
			fmt.Printf("(*) Pod updated: %s/%s\n", pod.Namespace, pod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("(-) Pod deleted: %s/%s\n", pod.Namespace, pod.Name)
		},
	})

	// SHARED ASPECT: Second handler on the SAME informer instance
	// Both handlers share the same watch connection and cache
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[SECOND-CONTROLLER] Also saw pod: %s/%s\n", pod.Namespace, pod.Name)
		},
	})

	// When a pod changes, BOTH handlers get notified from the same event stream
	// Only ONE HTTP connection is used for both handlers (efficient!)
	<-stopCh
}
