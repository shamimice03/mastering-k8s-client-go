package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func podStatus(clientset *kubernetes.Clientset) {
	for {
		pods, _ := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		for _, pod := range pods.Items {
			fmt.Printf("%s: %s\n", pod.Name, pod.Status.Phase)
		}
		fmt.Println("---")
	}
}

func main() {
	home, _ := os.UserHomeDir()
	config, _ := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube/config"))
	clientset, _ := kubernetes.NewForConfig(config)

	podStatus(clientset)
}
