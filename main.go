package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file")
	flag.Parse()

	// Use the provided kubeconfig file if specified, otherwise use the default in-cluster config
	var config *rest.Config
	var err error
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building kubeconfig: %s\n", err.Error())
			os.Exit(1)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building in-cluster kubeconfig: %s\n", err.Error())
			os.Exit(1)
		}
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

	// WaitGroup to wait for all Goroutines to finish
	var wg sync.WaitGroup

	// Channel to receive CRDs from Goroutines
	crdCh := make(chan *metav1.APIResourceList)

	// Retrieve CRDs concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		crds, err := crdClientset.Discovery().ServerPreferredResources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing CRDs: %s\n", err.Error())
			return
		}
		crdCh <- crds
	}()

	// Process CRDs concurrently
	for i := 0; i < 5; i++ { // Adjust the number of Goroutines as needed
		wg.Add(1)
		go func() {
			defer wg.Done()
			for crds := range crdCh {
				for _, apiResourceList := range crds {
					fmt.Printf("\nAPI Group: %s\n", apiResourceList.GroupVersion)
					for _, apiResource := range apiResourceList.APIResources {
						fmt.Printf("  - Kind: %s, Name: %s, Namespaced: %t\n", apiResource.Kind, apiResource.Name, apiResource.Namespaced)
					}
				}
			}
		}()
	}

	// Wait for all Goroutines to finish
	wg.Wait()

	// List all resources in the cluster
	resources, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing resources: %s\n", err.Error())
		os.Exit(1)
	}

	// Print the resources
	fmt.Println("\nAll Resources in the Cluster:")
	for _, apiResourceList := range resources {
		fmt.Printf("\nAPI Group: %s\n", apiResourceList.GroupVersion)
		for _, apiResource := range apiResourceList.APIResources {
			fmt.Printf("  - Kind: %s, Name: %s, Namespaced: %t\n", apiResource.Kind, apiResource.Name, apiResource.Namespaced)
		}
	}
}
