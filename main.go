package main

import (
	"flag"
	"fmt"
	"os"

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

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating clientset: %s\n", err.Error())
		os.Exit(1)
	}

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
