# k8s-restart-script

A simple Go script to restart all Deployments, StatefulSets, and DaemonSets in your current Kubernetes namespace whose names contain the substring `database`.  
It works by patching a restart annotation, similar to `kubectl rollout restart`.

---

## Features

- Restarts all matching Deployments, StatefulSets, and DaemonSets in the current namespace
- Uses your local kubeconfig for authentication (like `kubectl`)
- Prints a summary of which controllers were restarted

---

## Usage

```bash
go run k8-restart-script.go
```

- By default, it looks for controllers with `database` in their name in your current namespace.

---

## How it Works

- The script connects to your Kubernetes cluster using your kubeconfig.
- It lists all Deployments, StatefulSets, and DaemonSets in the current namespace.
- For each controller whose name contains `database`, it patches the pod template annotation to trigger a rolling restart.

---

## Requirements

- Go 1.18 or newer
- Access to a Kubernetes cluster (with a valid kubeconfig at `~/.kube/config`)

---

## Example Output

```
Restarted deployment: database-app
Restarted statefulset: database-store
Restarted daemonset: database-sidecar
```

---

## Customization

- To change the substring, edit the `match := "database"` line in the script.
- To use a different namespace, switch your current context or modify the script to accept a namespace argument.

---

## Build

```bash
go mod tidy
go build -o k8-restart
```

---

## Notes

- This script is intended for simple use cases and learning purposes.
- For production use, consider adding error handling, logging, and command-line flags for customization.