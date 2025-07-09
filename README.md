# k8s-restart-database

A Go CLI tool that gracefully restarts Kubernetes workloads (Deployments, StatefulSets, DaemonSets) containing a specific substring (default: `database`) in their names — similar to `kubectl rollout restart`.

---

## ✅ Features

- Mimics `kubectl rollout restart` by patching a restart annotation
- Supports `deployment`, `statefulset`, and `daemonset` controllers
- Allows namespace and match substring to be customized
- Provides a full summary of successful and failed restarts
- Uses local kubeconfig or in-cluster config for authentication

---

## 🚀 Usage

```bash
go run main.go [--namespace=NAMESPACE] [--controller=TYPE] [--match=STRING]
```

### 🔧 Flags

| Flag           | Description                                                                 |
|----------------|-----------------------------------------------------------------------------|
| `--namespace`  | Namespace to operate in. Defaults to current context's namespace            |
| `--controller` | Controller type to restart: `deployment`, `statefulset`, `daemonset`        |
| `--match`      | Substring to match in controller names (case-sensitive). Defaults to `database` |

---

## 🧪 Examples

### Restart all controllers in the current namespace matching `"database"`
```bash
go run main.go
```

### Restart only StatefulSets containing `"postgres"` in the `prod` namespace
```bash
go run main.go --namespace=prod --controller=statefulset --match=postgres
```

---

## 🔐 Authentication

- Uses `~/.kube/config` by default (like `kubectl`)
- Falls back to in-cluster config if running inside a Pod

---

## 📦 Build

```bash
go mod tidy
go build -o k8s-restart
```

---

## 🛑 Exit Codes

- `0` — all restarts succeeded
- `1` — at least one restart failed (details shown in summary)
