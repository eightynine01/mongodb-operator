# MongoDB Operator for Kubernetes

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.26+-326CE5?logo=kubernetes)](https://kubernetes.io/)

A Kubernetes Operator for deploying and managing MongoDB ReplicaSets and Sharded Clusters.

## Overview

MongoDB Operator automates the deployment, scaling, and management of MongoDB clusters on Kubernetes. It provides a declarative way to manage MongoDB infrastructure using Custom Resource Definitions (CRDs).

### Features

- **MongoDB ReplicaSet**: Deploy highly available 3+ member replica sets with automatic failover
- **Sharded Cluster**: Deploy distributed clusters with config servers, shards, and mongos routers
- **TLS Encryption**: Automatic TLS certificate management with cert-manager integration
- **Authentication**: SCRAM-SHA-256 authentication with keyfile support for internal cluster communication
- **Monitoring**: Prometheus metrics export with ServiceMonitor support
- **Backup/Restore**: Automated backups to S3-compatible storage or PVC
- **Auto-scaling**: Horizontal Pod Autoscaler support for Mongos routers

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    MongoDB Operator                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  MongoDB    │  │ MongoDBShar │  │    MongoDBBackup        │  │
│  │  Controller │  │ Controller  │  │    Controller           │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
│         │                │                      │                │
│         ▼                ▼                      ▼                │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                  Resource Builder                           ││
│  │  (StatefulSets, Deployments, Services, Secrets, Jobs)       ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐        │
│  │  StatefulSet  │  │  StatefulSet  │  │  Deployment   │        │
│  │  (ReplicaSet) │  │  (Shards)     │  │  (Mongos)     │        │
│  └───────────────┘  └───────────────┘  └───────────────┘        │
│                                                                  │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐        │
│  │   Services    │  │    Secrets    │  │  ConfigMaps   │        │
│  └───────────────┘  └───────────────┘  └───────────────┘        │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Kubernetes cluster v1.26+
- Helm v3.8+
- kubectl configured with cluster access

### Installation

```bash
# Add Helm repository
helm repo add mongodb-operator https://eightynine01.github.io/mongodb-operator
helm repo update

# Install the operator
helm install mongodb-operator mongodb-operator/mongodb-operator \
  --namespace mongodb-operator-system \
  --create-namespace
```

### Deploy a MongoDB ReplicaSet

```yaml
apiVersion: mongodb.keiailab.com/v1alpha1
kind: MongoDB
metadata:
  name: my-mongodb
  namespace: database
spec:
  members: 3
  version:
    version: "8.2"
  storage:
    storageClassName: standard
    size: 10Gi
  auth:
    mechanism: SCRAM-SHA-256
    adminCredentialsSecretRef:
      name: mongodb-admin
  monitoring:
    enabled: true
```

```bash
# Create namespace and credentials
kubectl create namespace database
kubectl create secret generic mongodb-admin \
  --from-literal=username=admin \
  --from-literal=password=your-secure-password \
  -n database

# Deploy MongoDB
kubectl apply -f mongodb-replicaset.yaml
```

### Deploy a Sharded Cluster

```yaml
apiVersion: mongodb.keiailab.com/v1alpha1
kind: MongoDBSharded
metadata:
  name: my-sharded
  namespace: database
spec:
  version:
    version: "8.2"
  configServer:
    members: 3
    storage:
      size: 5Gi
  shards:
    count: 3
    membersPerShard: 3
    storage:
      size: 50Gi
  mongos:
    replicas: 2
    service:
      type: LoadBalancer
```

## Custom Resource Definitions

### MongoDB (ReplicaSet)

| Field | Description | Default |
|-------|-------------|---------|
| `spec.members` | Number of replica set members | `3` |
| `spec.version.version` | MongoDB version | `8.2` |
| `spec.storage.storageClassName` | Storage class name | - |
| `spec.storage.size` | PVC size per member | `10Gi` |
| `spec.auth.mechanism` | Authentication mechanism | `SCRAM-SHA-256` |
| `spec.tls.enabled` | Enable TLS | `false` |
| `spec.monitoring.enabled` | Enable Prometheus metrics | `false` |
| `spec.arbiter.enabled` | Enable arbiter node | `false` |

### MongoDBSharded

| Field | Description | Default |
|-------|-------------|---------|
| `spec.configServer.members` | Config server replica count | `3` |
| `spec.shards.count` | Number of shards | `2` |
| `spec.shards.membersPerShard` | Members per shard | `3` |
| `spec.mongos.replicas` | Mongos router replicas | `2` |
| `spec.mongos.autoScaling.enabled` | Enable HPA for mongos | `false` |

### MongoDBBackup

| Field | Description | Default |
|-------|-------------|---------|
| `spec.clusterRef.name` | Target cluster name | - |
| `spec.clusterRef.kind` | Target cluster kind | `MongoDB` |
| `spec.type` | Backup type (full/incremental) | `full` |
| `spec.compression` | Enable compression | `true` |
| `spec.storage.type` | Storage type (s3/pvc) | `s3` |

## Configuration

### TLS with cert-manager

```yaml
spec:
  tls:
    enabled: true
    certManager:
      issuerRef:
        name: letsencrypt-prod
        kind: ClusterIssuer
```

### Prometheus Monitoring

```yaml
spec:
  monitoring:
    enabled: true
    prometheusRule:
      enabled: true
    serviceMonitor:
      interval: 30s
```

### Backup to S3

```yaml
apiVersion: mongodb.keiailab.com/v1alpha1
kind: MongoDBBackup
metadata:
  name: daily-backup
spec:
  clusterRef:
    name: my-mongodb
    kind: MongoDB
  storage:
    type: s3
    s3:
      bucket: mongodb-backups
      endpoint: https://s3.amazonaws.com
      region: us-east-1
      credentialsRef:
        name: s3-credentials
```

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Kind or Minikube (for local testing)

### Building

```bash
# Build the operator
make build

# Run tests
make test

# Build Docker image
make docker-build IMG=your-registry/mongodb-operator:tag

# Push Docker image
make docker-push IMG=your-registry/mongodb-operator:tag
```

### Local Development

```bash
# Install CRDs
make install

# Run the operator locally
make run

# Create a sample MongoDB
kubectl apply -f config/samples/mongodb_replicaset.yaml
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

### Third-Party Licenses

This operator manages MongoDB databases but does not include or distribute MongoDB software. MongoDB Community Server is licensed under the [Server Side Public License (SSPL)](https://www.mongodb.com/licensing/server-side-public-license).

**Important License Notes:**
- This operator (Apache 2.0) is independent software that orchestrates MongoDB deployments
- MongoDB container images are pulled from official MongoDB repositories
- Users are responsible for complying with MongoDB's licensing terms
- The operator does not modify or redistribute MongoDB binaries

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## Support

- **Issues**: [GitHub Issues](https://github.com/eightynine01/mongodb-operator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/eightynine01/mongodb-operator/discussions)

## Roadmap

- [ ] Point-in-Time Recovery (PITR)
- [ ] Automated version upgrades
- [ ] Cross-cluster replication
- [ ] Grafana dashboard templates
- [ ] Backup scheduling with CronJob

## Acknowledgments

- [Kubernetes](https://kubernetes.io/)
- [Operator SDK](https://sdk.operatorframework.io/)
- [MongoDB](https://www.mongodb.com/)
- [Bitnami MongoDB Charts](https://github.com/bitnami/charts) for inspiration
