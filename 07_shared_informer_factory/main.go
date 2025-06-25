package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1" // This imports Pod, Service, etc.
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
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

func main() {
	clientset := createClientset()

	//Single shared factory
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	// Controllers using same factory
	startPodMonitor(factory)
	startDeploymentManager(factory)
	startPodScaler(factory)

	// Start all informers at once
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)
	<-stopCh

}

func startPodMonitor(factory informers.SharedInformerFactory) {
	podInformer := factory.Core().V1().Pods() // Gets shared Pod informer

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[Monitor] Pod added: %s\n", pod.Name)
		},
	})
}

// Deployment Manager by implementing cache.ResourceEventHandler
type DeploymentHandler struct{}

func (h *DeploymentHandler) OnAdd(obj interface{}, isInInitialList bool) {
	deployment := obj.(*appsv1.Deployment)
	fmt.Printf("[Manager] Deployment added: %s\n", deployment.Name)
}

func (h *DeploymentHandler) OnUpdate(oldObj, newObj interface{}) {
	// Empty implementation
}

func (h *DeploymentHandler) OnDelete(obj interface{}) {
	// Empty implementation
}

func startDeploymentManager(factory informers.SharedInformerFactory) {
	deploymentInformer := factory.Apps().V1().Deployments()

	handler := &DeploymentHandler{}
	deploymentInformer.Informer().AddEventHandler(handler)

}

// func startDeploymentManager(factory informers.SharedInformerFactory) {
//     deploymentInformer := factory.Apps().V1().Deployments() // Different resource type

//     deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
//         AddFunc: func(obj interface{}) {
//             deployment := obj.(*appsv1.Deployment)
//             fmt.Printf("[Manager] Deployment added: %s\n", deployment.Name)
//         },
//     })
// }

// Controller 1: Pod Monitor
// func startPodMonitor(factory informers.SharedInformerFactory) {
// 	podInformer := factory.Core().V1().Pods() // Gets shared Pod informer

// 	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
// 		AddFunc: func(obj interface{}) {
// 			pod := obj.(*corev1.Pod)
// 			fmt.Printf("[Monitor] Pod added: %s\n", pod.Name)
// 		},
// 	})
// }

// // Controller 2: Pod Scaler (uses SAME Pod informer as Controller 1)
func startPodScaler(factory informers.SharedInformerFactory) {
	podInformer := factory.Core().V1().Pods() // Gets the SAME shared Pod informer

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*corev1.Pod)
			fmt.Printf("[Scaler] Pod updated: %s\n", pod.Name)
			// Scaling logic here
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[Scaler] Pod deleted: %s\n", pod.Name)
		},
	})
}
