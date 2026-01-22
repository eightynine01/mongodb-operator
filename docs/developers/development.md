# Development Guide

## Overview

This guide helps developers set up a local development environment for the MongoDB Operator, run the operator locally, and contribute effectively.

## Local Development Setup

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Kind or Minikube (recommended for local testing)
- Make (for build automation)

### Setting Up Kind Cluster

```bash
# Install Kind
go install sigs.k8s.io/kind@v0.20.0

# Create Kind cluster
kind create cluster --name mongodb-operator-dev

# Verify cluster
kubectl cluster-info
kubectl get nodes
```

**Kind Configuration with Extra Ports** (optional, for LoadBalancer services):

```yaml
# kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30080
        hostPort: 80
      - containerPort: 30443
        hostPort: 443
```

```bash
kind create cluster --config kind-config.yaml --name mongodb-operator-dev
```

### Alternative: Minikube Setup

```bash
# Install Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64
sudo install minikube-darwin-amd64 /usr/local/bin/minikube

# Start Minikube
minikube start --cpus=4 --memory=8192

# Enable ingress
minikube addons enable ingress
```

## Running Operator Locally

### Method 1: Using Make (Recommended)

```bash
# Install CRDs
make install

# Run operator locally
make run

# In another terminal, create test resources
kubectl apply -f config/samples/mongodb_replicaset.yaml
```

### Method 2: Direct Go Run

```bash
# Install CRDs first
kubectl apply -f config/crd/bases/mongodb.keiailab.com_mongodbs.yaml
kubectl apply -f config/crd/bases/mongodb.keiailab.com_mongodbshardeds.yaml

# Set Kubernetes config
export KUBECONFIG=~/.kube/config

# Run operator
go run ./main.go
```

### Method 3: Deploy to Cluster

```bash
# Build Docker image
make docker-build IMG=mongodb-operator:dev

# Load image into Kind
kind load docker-image mongodb-operator:dev --name mongodb-operator-dev

# Deploy operator
kubectl apply -f config/manager/manager.yaml

# Update deployment to use dev image
kubectl set image deployment/mongodb-operator controller=mongodb-operator:dev -n mongodb-operator-system
```

## Hot Reload During Development

### Using Air (Live Reload)

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Create .air.toml in project root
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ."
  bin = "tmp/main"
  include_ext = ["go", "yaml"]
  exclude_dir = ["tmp", "vendor"]
  delay = 1000

[log]
  time = true

# Run with Air
air
```

### Using entr (Alternative)

```bash
# Install entr
brew install entr

# Watch for changes and rebuild
find . -name "*.go" | entr -r sh -c 'make build && ./bin/manager'
```

## Debugging Operator

### Using Delve (Go Debugger)

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run operator with Delve
dlv debug ./main.go --listen=:2345 --headless=true --api-version=2 --accept-multiclient

# Connect with VS Code
# Create .vscode/launch.json:
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Connect to Delve",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "remotePath": "${workspaceFolder}",
      "port": 2345,
      "host": "127.0.0.1"
    }
  ]
}
```

### Using kubectl Debug

```bash
# Debug operator pod
kubectl debug -n mongodb-operator-system -it $(kubectl get pods -n mongodb-operator-system -o name | head -1) --image=busybox --target=manager

# View operator logs in real-time
kubectl logs -n mongodb-operator-system -f -l app.kubernetes.io/name=mongodb-operator
```

### Adding Debug Logging

```go
// In your controller code
import (
    "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    log.Info("Reconciling MongoDB", "name", req.Name)

    // Enable debug logging
    log.V(1).Info("Detailed debug information", "key", "value")

    // Your reconciliation logic...

    return ctrl.Result{}, nil
}
```

Set log level:
```bash
kubectl set env deployment/mongodb-operator LOG_LEVEL=debug -n mongodb-operator-system
```

## Code Structure Overview

```
mongodb-operator/
├── api/
│   └── v1alpha1/
│       ├── mongodb_types.go       # MongoDB CRD types
│       ├── mongodbsharded_types.go # MongoDBSharded CRD types
│       ├── mongodbbackup_types.go  # MongoDBBackup CRD types
│       ├── groupversion_info.go    # API version info
│       └── zz_generated.deepcopy.go # Generated deepcopy methods
├── controllers/
│   ├── mongodb_controller.go      # ReplicaSet controller
│   ├── mongodbsharded_controller.go # Sharded cluster controller
│   ├── mongodbbackup_controller.go  # Backup controller
│   └── suite_test.go              # Controller tests
├── internal/
│   ├── mongodb/                   # MongoDB package
│   │   ├── executor.go            # MongoDB command executor
│   │   ├── replicaset.go          # ReplicaSet operations
│   │   ├── auth.go                # Authentication handling
│   │   ├── sharding.go            # Sharding operations
│   │   └── metrics.go             # Metrics collection
│   ├── resource/
│   │   ├── builder.go             # Resource builder
│   │   ├── statefulset.go         # StatefulSet builder
│   │   ├── deployment.go          # Deployment builder
│   │   ├── service.go             # Service builder
│   │   └── secret.go              # Secret builder
│   └── config/
│       └── config.go              # Configuration handling
├── config/
│   ├── crd/                       # CRD definitions
│   ├── rbac/                      # RBAC rules
│   ├── manager/                   # Manager deployment
│   └── samples/                   # Sample manifests
├── main.go                        # Operator entry point
└── Makefile                       # Build automation
```

## Development Workflow

### Making Changes

```bash
# 1. Create feature branch
git checkout -b feature/my-feature

# 2. Make changes to code
vim controllers/mongodb_controller.go

# 3. Test locally
make install
make run

# 4. Apply test manifest
kubectl apply -f config/samples/mongodb_replicaset.yaml

# 5. Verify changes
kubectl get pods -w
kubectl logs -n mongodb-operator-system -f

# 6. Run tests
make test
```

### Adding New CRD Fields

1. Update type definition in `api/v1alpha1/mongodb_types.go`:
   ```go
   type MongoDBSpec struct {
       // ... existing fields ...
       MyNewField string `json:"myNewField,omitempty"`
   }
   ```

2. Generate deepcopy methods:
   ```bash
   make generate
   ```

3. Update controller to handle new field
4. Update documentation and samples

### Adding Custom Metrics

```go
// In your controller
import (
    "github.com/prometheus/client_golang/prometheus"
    "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
    customMetric = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mongodb_operator_custom_metric",
            Help: "Custom metric for tracking",
        },
        []string{"name"},
    )
)

func init() {
    metrics.Registry.MustRegister(customMetric)
}

// Use metric in controller
customMetric.WithLabelValues(req.Name).Inc()
```

## Testing Changes

```bash
# Run unit tests
make test

# Run specific test
go test ./controllers -v -run TestMongoDBReconciler

# Run integration tests
make test-integration

# Test with Kind
make test-kind

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Common Development Tasks

### Rebuilding CRDs

```bash
# Regenerate CRD manifests
make manifests

# Install updated CRDs
make install
```

### Regenerating Code

```bash
# Generate deepcopy methods
make generate

# Generate manifests (CRDs, RBAC)
make manifests

# Generate mock files (if using mocks)
make mocks
```

### Updating Dependencies

```bash
# Update go.mod
go mod tidy

# Update specific dependency
go get github.com/pkg/errors@latest

# Vendor dependencies
go mod vendor
```

## Troubleshooting Development Issues

### CRD Changes Not Reflecting

```bash
# Delete and reinstall CRDs
kubectl delete crd mongodbs.mongodb.keiailab.com
make install

# Check CRD status
kubectl get crd mongodbs.mongodb.keiailab.com -o yaml
```

### Operator Not Reconciling

```bash
# Check if CRD is recognized
kubectl get crd | grep mongodb

# Check operator logs for errors
kubectl logs -n mongodb-operator-system -f

# Verify RBAC permissions
kubectl auth can-i create mongodbs -n database --as=system:serviceaccount:mongodb-operator-system:mongodb-operator
```

### Make Errors

```bash
# Clean build artifacts
make clean

# Rebuild from scratch
make build
```
