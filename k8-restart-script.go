package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "strings" // used to match the name of the controller with the string "database"
    "time"

    "k8s.io/apimachinery/pkg/types" //used for starategic merge patch object so I don't have to send full payload. 
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
    // Get current namespace from kubeconfig
    rules := clientcmd.NewDefaultClientConfigLoadingRules()
    cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
    namespace, _, err := cfg.Namespace()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error resolving namespace: %v\n", err)
        os.Exit(1)
    }

    // use the current context in kubeconfig
    config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating Kubernetes config: %v\n", err)
        os.Exit(1)
    }
	// create the clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
        os.Exit(1)
    }

    match := "database"

    // Restart Deployments if any match the string "database"
    deployments, _ := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
    for _, d := range deployments.Items {
        if strings.Contains(d.Name, match) {
            restartController(clientset, namespace, "deployment", d.Name)
        }
    }

    // Restart StatefulSets if any match the string "database"
    stateful, _ := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
    for _, s := range stateful.Items {
        if strings.Contains(s.Name, match) {
            restartController(clientset, namespace, "statefulset", s.Name)
        }
    }

    // Restart DaemonSets if any match the string "database"
    daemons, _ := clientset.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
    for _, d := range daemons.Items {
        if strings.Contains(d.Name, match) {
            restartController(clientset, namespace, "daemonset", d.Name)
        }
    }
}
func restartController(clientset *kubernetes.Clientset, namespace, kind, name string) {
    patch := map[string]interface{}{
        "spec": map[string]interface{}{
            "template": map[string]interface{}{
                "metadata": map[string]interface{}{
                    "annotations": map[string]string{
                        "kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
                    },
                },
            },
        },
    }
    bytes, _ := json.Marshal(patch)

    switch kind { // Makes the correct call based on type of managed pods and print if the restart was successful for each kind of controller
    case "deployment":
        _, err := clientset.AppsV1().Deployments(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, bytes, metav1.PatchOptions{})
        if err != nil {
            fmt.Printf("Failed to restart deployment %s: %v\n", name, err)
        } else {
            fmt.Printf("Restarted deployment: %s\n", name)
        }
    case "statefulset":
        _, err := clientset.AppsV1().StatefulSets(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, bytes, metav1.PatchOptions{})
        if err != nil {
            fmt.Printf("Failed to restart statefulset %s: %v\n", name, err)
        } else {
            fmt.Printf("Restarted statefulset: %s\n", name)
        }
    case "daemonset":
        _, err := clientset.AppsV1().DaemonSets(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, bytes, metav1.PatchOptions{})
        if err != nil {
            fmt.Printf("Failed to restart daemonset %s: %v\n", name, err)
        } else {
            fmt.Printf("Restarted daemonset: %s\n", name)
        }
    }
}