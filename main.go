package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "sort"
    "sync"

    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

// PodMetrics represents metrics for a pod
type PodMetrics struct {
    Pod     metav1.Pod
    Metrics *metav1.PodMetrics
}

// getResourceUsage returns the resource usage for a pod (CPU or memory)
func getResourceUsage(metrics *metav1.PodMetrics, resourceType string) resource.Quantity {
    switch resourceType {
    case "cpu":
        return metrics.Containers[0].Usage["cpu"]
    case "memory":
        return metrics.Containers[0].Usage["memory"]
    default:
        return resource.Quantity{}
    }
}

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

    defer clientset.Close() // Close clientset on exit

    // Retrieve all pods
    pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error listing pods: %s\n", err.Error())
        os.Exit(1)
    }

    // WaitGroup to wait for all Goroutines to finish
    var wg sync.WaitGroup

    // Channel to receive pod metrics from Goroutines
    metricsCh := make(chan PodMetrics)

    // Retrieve resource usage metrics for each pod concurrently
    for _, pod := range pods.Items {
        wg.Add(1)
        go func(pod metav1.Pod) {
            defer wg.Done()
            metricsGetter := clientset.CoreV1().Pods(pod.Namespace)
            metrics, err := metricsGetter.GetMetrics(context.Background(), pod.Name, metav1.GetOptions{})
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error getting metrics for pod %s/%s: %s\n", pod.Namespace, pod.Name, err.Error())
                return
            }
            metricsCh <- PodMetrics{Pod: pod, Metrics: metrics}
        }(pod)
    }

    // Close the metrics channel after all Goroutines finish
    go func() {
        wg.Wait()
        close(metricsCh)
    }()

    // Collect pod metrics from the channel
    var podMetrics []PodMetrics
    for metric := range metricsCh {
        podMetrics = append(podMetrics, metric)
    }

    // Sort pods by CPU and memory usage
    sort.Slice(podMetrics, func(i, j int) bool {
        return getResourceUsage(podMetrics[i].Metrics, "cpu").Cmp(getResourceUsage(podMetrics[j].Metrics, "cpu")) > 0 ||
            getResourceUsage(pod)
	},
