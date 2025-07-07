package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	// Create client
	clientset := createClientSet()

	// Single factory for all informers
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	// Setup multiple informers using same factory
	setupPodMonitor(factory)
	setupDeploymentMonitor(factory)
	setupPodUpdateMonitor(factory)

	// Start all informers at once
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)
	<-stopCh
}

// Controller 1: Pod Monitor
func setupPodMonitor(factory informers.SharedInformerFactory) {
	podInformer := factory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[Monitor] Pod added: %s\n", pod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[Monitor] Pod deleted: %s\n", pod.Name)
		},
	})
}

// Controller 2: Deployment Manager
// Deployment Manager by implementing cache.ResourceEventHandler
type DeploymentHandler struct{}

func (h *DeploymentHandler) OnAdd(obj interface{}, isInInitialList bool) {
	deployment := obj.(*appsv1.Deployment)
	fmt.Printf("[Manager] Deployment added: %s\n", deployment.Name)
}

func (h *DeploymentHandler) OnUpdate(oldObj, newObj interface{}) {
	// Implementation
}

func (h *DeploymentHandler) OnDelete(obj interface{}) {
	// Implementation
}

func setupDeploymentMonitor(factory informers.SharedInformerFactory) {
	deploymentInformer := factory.Apps().V1().Deployments()

	handler := &DeploymentHandler{}
	deploymentInformer.Informer().AddEventHandler(handler)

}

// Controller 3: Pod Update Monitor (uses SAME Pod informer as Controller 1)
func setupPodUpdateMonitor(factory informers.SharedInformerFactory) {
	podInformer := factory.Core().V1().Pods() // Gets the SAME shared Pod informer

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*corev1.Pod)
			fmt.Printf("[PodUpdateMonitor] Pod updated: %s\n", pod.Name)
			// logic here
		},
	})
}
