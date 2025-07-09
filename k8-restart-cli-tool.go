package main


import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "strings"
    "time"


    "k8s.io/apimachinery/pkg/types"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


var ( // Command line flags
    nsFlag         = flag.String("namespace", "", "Namespace to operate in. Defaults to current context's namespace.")
    controllerFlag = flag.String("controller", "", "Controller type to target: deployment, statefulset, daemonset. Defaults to all.")
    matchFlag      = flag.String("match", "database", "Substring to match in controller names.")
)


var ( // Global variables to track state
    hadErrors           = false
    succeededControllers []string
    failedControllers    = make(map[string]string)
)


func main() {
    flag.Parse()


    namespace, err := resolveNamespace(*nsFlag)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error resolving namespace: %v\n", err)
        os.Exit(1)
    }


    clientset, err := getKubernetesClient()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
        os.Exit(1)
    }


    typesToCheck := []string{"deployment", "statefulset", "daemonset"}
    if *controllerFlag != "" {
        if !contains(typesToCheck, strings.ToLower(*controllerFlag)) {
            fmt.Fprintf(os.Stderr, "Invalid controller type: %s\n", *controllerFlag)
            os.Exit(1)
        }
        typesToCheck = []string{strings.ToLower(*controllerFlag)}
    }


    for _, controllerType := range typesToCheck {
        switch controllerType {
        case "deployment":
            restartDeployments(clientset, namespace)
        case "statefulset":
            restartStatefulSets(clientset, namespace)
        case "daemonset":
            restartDaemonSets(clientset, namespace)
        }
    }


    // Print summary
    fmt.Println("\n--- Restart Summary ---")
    fmt.Printf("Successful restarts: %d\n", len(succeededControllers))
    for _, name := range succeededControllers {
        fmt.Printf("✓ %s\n", name)
    }
    if len(failedControllers) > 0 {
        fmt.Printf("\nFailed restarts: %d\n", len(failedControllers))
        for name, reason := range failedControllers {
            fmt.Printf("✗ %s - %s\n", name, reason)
        }
    }


    if hadErrors {
        os.Exit(1)
    }
}


func resolveNamespace(nsFlag string) (string, error) {
    if nsFlag != "" {
        return nsFlag, nil
    }


    rules := clientcmd.NewDefaultClientConfigLoadingRules()
    cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
    ns, _, err := cfg.Namespace()
    return ns, err
}


func getKubernetesClient() (*kubernetes.Clientset, error) {
    config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
    if err != nil {
        return nil, err
    }
    return kubernetes.NewForConfig(config)
}


func restartDeployments(clientset *kubernetes.Clientset, namespace string) {
    deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error listing deployments: %v\n", err)
        hadErrors = true
        return
    }
    for _, d := range deployments.Items {
        if strings.Contains(d.Name, *matchFlag) {
            fmt.Printf("Restarting deployment: %s\n", d.Name)
            restartController(clientset, namespace, "deployment", d.Name)
        }
    }
}


func restartStatefulSets(clientset *kubernetes.Clientset, namespace string) {
    sets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error listing statefulsets: %v\n", err)
        hadErrors = true
        return
    }
    for _, s := range sets.Items {
        if strings.Contains(s.Name, *matchFlag) {
            fmt.Printf("Restarting statefulset: %s\n", s.Name)
            restartController(clientset, namespace, "statefulset", s.Name)
        }
    }
}


func restartDaemonSets(clientset *kubernetes.Clientset, namespace string) {
    daemons, err := clientset.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error listing daemonsets: %v\n", err)
        hadErrors = true
        return
    }
    for _, d := range daemons.Items {
        if strings.Contains(d.Name, *matchFlag) {
            fmt.Printf("Restarting daemonset: %s\n", d.Name)
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


    switch kind {
    case "deployment":
        _, err := clientset.AppsV1().Deployments(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, bytes, metav1.PatchOptions{})
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error restarting deployment %s: %v\n", name, err)
            hadErrors = true
            failedControllers["deployment/"+name] = err.Error()
        } else {
            succeededControllers = append(succeededControllers, "deployment/"+name)
        }
    case "statefulset":
        _, err := clientset.AppsV1().StatefulSets(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, bytes, metav1.PatchOptions{})
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error restarting statefulset %s: %v\n", name, err)
            hadErrors = true
            failedControllers["statefulset/"+name] = err.Error()
        } else {
            succeededControllers = append(succeededControllers, "statefulset/"+name)
        }
    case "daemonset":
        _, err := clientset.AppsV1().DaemonSets(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, bytes, metav1.PatchOptions{})
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error restarting daemonset %s: %v\n", name, err)
            hadErrors = true
            failedControllers["daemonset/"+name] = err.Error()
        } else {
            succeededControllers = append(succeededControllers, "daemonset/"+name)
        }
    }
}


func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false