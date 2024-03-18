package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file")
	flag.Parse()

	// Use the current context in kubeconfig file
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %s\n", err.Error())
		os.Exit(1)
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating clientset: %s\n", err.Error())
		os.Exit(1)
	}

	// Create a CRD clientset
	crdClientset, err := clientset.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CRD clientset: %s\n", err.Error())
		os.Exit(1)
	}

	crds, err := crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing CRDs: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("Custom Resource Definitions:")
	for _, crd := range crds.Items {
		fmt.Printf("- %s\n", crd.Name)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing namespaces: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("Namespaces:")
	for _, ns := range namespaces.Items {
		fmt.Printf("- %s\n", ns.Name)
	}

	for _, ns := range namespaces.Items {
		fmt.Printf("\nResources in Namespace: %s\n", ns.Name)

		resources, err := clientset.Discovery().ServerPreferredResources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing resources: %s\n", err.Error())
			os.Exit(1)
		}

		for _, apiResourceList := range resources {
			fmt.Printf("API Group: %s\n", apiResourceList.GroupVersion)
			for _, apiResource := range apiResourceList.APIResources {
				fmt.Printf("  - Kind: %s, Name: %s, Namespaced: %t\n", apiResource.Kind, apiResource.Name, apiResource.Namespaced)
			}
		}
	}
}
