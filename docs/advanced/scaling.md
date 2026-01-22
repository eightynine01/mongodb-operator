# Scaling Strategies

## Overview

MongoDB Operator supports horizontal and vertical scaling for both ReplicaSets and Sharded Clusters. This guide covers scaling operations and best practices.

## Horizontal Scale Out (Adding Shards)

The operator automatically handles shard scaling when you increase `spec.shards.count`.

### Scale Out Process

When increasing shard count, the operator performs:

1. Creates new Shard StatefulSet and headless Service
2. Waits for all pods to become ready
3. Initializes the new shard's ReplicaSet (`rs.initiate()`)
4. Registers the new shard with mongos (`sh.addShard()`)
5. MongoDB balancer automatically migrates chunks to the new shard

### Example: Scale from 3 to 5 Shards

```bash
# Check current shard count
kubectl get mongodbsharded my-cluster -o jsonpath='{.spec.shards.count}'
# Output: 3

# Scale out to 5 shards
kubectl patch mongodbsharded my-cluster --type='merge' \
  -p '{"spec":{"shards":{"count":5}}}'
```

### Monitor Scaling Progress

```bash
# Watch new shard pods
kubectl get pods -l app.kubernetes.io/component=shard -w

# Check shard status
kubectl get mongodbsharded my-cluster -o yaml | grep -A 20 status:

# Verify shards registered with mongos
kubectl exec -it my-cluster-mongos-0 -n database -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.status()'
```

### Status Tracking

```yaml
status:
  shardsInitialized: [true, true, true, true, true]
  shardsAdded: [true, true, true, true, true]
  shards:
    - name: my-cluster-shard-0
      phase: Running
    - name: my-cluster-shard-1
      phase: Running
    - name: my-cluster-shard-2
      phase: Running
    - name: my-cluster-shard-3
      phase: Running
    - name: my-cluster-shard-4
      phase: Running
```

### Managing Balancer During Scale

The MongoDB balancer automatically redistributes data. Monitor and configure:

```bash
# Check balancer state
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.getBalancerState()'

# Temporarily disable balancer
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.stopBalancer()'

# Re-enable balancer
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.startBalancer()'

# Configure balancer window
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'db.settings.update({_id: "balancer"}, {$set: {activeWindow: {start: "23:00", stop: "06:00"}}}, {upsert: true})'
```

## Vertical Scaling (Resources)

Update resource requests/limits triggers a rolling restart.

### Scaling ReplicaSet Resources

```yaml
apiVersion: mongodb.keiailab.com/v1alpha1
kind: MongoDB
metadata:
  name: my-mongodb
  namespace: database
spec:
  members: 3
  resources:
    requests:
      memory: "4Gi"
      cpu: "2"
    limits:
      memory: "8Gi"
      cpu: "4"
  storage:
    size: 20Gi
```

### Scaling Sharded Cluster Resources

```bash
# Scale shard resources
kubectl patch mongodbsharded my-cluster --type='merge' -p '{
  "spec": {
    "shards": {
      "resources": {
        "requests": {"memory": "4Gi", "cpu": "2"},
        "limits": {"memory": "8Gi", "cpu": "4"}
      }
    }
  }
}'

# Scale config server resources
kubectl patch mongodbsharded my-cluster --type='merge' -p '{
  "spec": {
    "configServer": {
      "resources": {
        "requests": {"memory": "2Gi", "cpu": "1"},
        "limits": {"memory": "4Gi", "cpu": "2"}
      }
    }
  }
}'
```

### Scaling Storage

**Note**: Storage size increases are immutable for PVCs. You must create new PVCs and migrate data.

```yaml
# Cannot directly increase PVC size
# Instead, create a new cluster with larger storage and migrate
spec:
  storage:
    size: 50Gi  # Only works on new clusters
```

## Mongos Scaling

Scale mongos routers up or down for load distribution.

### Manual Scaling

```bash
# Scale up to 3 mongos replicas
kubectl patch mongodbsharded my-cluster --type='merge' \
  -p '{"spec":{"mongos":{"replicas":3}}}'

# Scale down to 1 mongos replica
kubectl patch mongodbsharded my-cluster --type='merge' \
  -p '{"spec":{"mongos":{"replicas":1}}}'
```

### Horizontal Pod Autoscaler (HPA)

Enable HPA for automatic mongos scaling:

```yaml
apiVersion: mongodb.keiailab.com/v1alpha1
kind: MongoDBSharded
metadata:
  name: my-cluster
  namespace: database
spec:
  mongos:
    replicas: 2
    autoScaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      targetCPUUtilizationPercentage: 70
      targetMemoryUtilizationPercentage: 80
```

```bash
# Manually create HPA (if not using autoScaling in CRD)
kubectl autoscale deployment my-cluster-mongos \
  --cpu-percent=70 \
  --min=2 \
  --max=10 \
  -n database

# Check HPA status
kubectl get hpa -n database
```

## Best Practices for Production Scaling

### Pre-Scaling Planning

1. **Monitor Performance Metrics**: Before scaling, identify bottlenecks
2. **Plan Shard Key Choice**: Critical for efficient data distribution
3. **Sufficient Storage**: Plan for data growth across shards
4. **Network Bandwidth**: Ensure adequate bandwidth for inter-shard communication

### Scaling Recommendations

| Component | Minimum | Production | Notes |
|-----------|---------|------------|-------|
| ReplicaSet Members | 3 | 3-7 | Odd numbers preferred |
| Config Server Members | 3 | 3 | Three required |
| Shards | 2 | 3-10 | Start with 3-5 |
| Mongos Replicas | 2 | 3-5 | Based on client load |

### Storage Sizing

**Formula:**
```
Required Storage = (Database Size × Replication Factor) × Growth Buffer
```

Example for 1TB database with 3-way replication:
```
Required Storage = (1TB × 3) × 1.5 = 4.5TB total
Per Shard = 4.5TB / Number of Shards
```

### Resource Sizing

**CPU:**
- 1-2 cores per shard member (workload dependent)
- Monitor `mongodb_cpu_user` and `mongodb_cpu_kernel` metrics

**Memory:**
- Minimum 512Mi per mongos (256Mi causes OOM)
- 2-4Gi per shard member for production
- Cache working set in RAM for optimal performance

**Storage I/O:**
- Use SSD storage for sharded clusters
- Minimum 3000 IOPS recommended
- Monitor `mongodb_storage_used_bytes` and latency metrics

## Scaling Limitations

### Currently Not Supported

1. **Shard Scale-In**: Removing shards requires manual intervention
   ```bash
   # Manual process
   kubectl exec -it my-cluster-mongos-0 -c mongos -- \
     mongosh -u admin -p $PASSWORD --eval 'sh.removeShard("shard-replica/shard-0.example.com:27018")'
   ```

2. **ReplicaSet Member Removal**: Cannot remove members automatically
   ```bash
   # Manual process
   kubectl exec -it my-mongodb-0 -c mongod -- \
     mongosh -u admin -p $PASSWORD --eval 'rs.remove("my-mongodb-2.my-mongodb:27017")'
   ```

3. **Storage Reduction**: PVC size increases are one-way
   - Cannot decrease PVC size
   - Cannot expand existing PVCs (requires migration)

### Scale-In Workarounds

**For Shards:**
1. Drain data from shard using balancer
2. Remove shard manually
3. Delete StatefulSet and associated resources

**For ReplicaSets:**
1. Step down primary
2. Remove member manually
3. Update CRD member count
4. Delete pod

## Scaling Example: Production Migration

Scenario: Migrate from 3 to 5 shards for production workload

```bash
# 1. Pre-migration: Monitor current metrics
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.status()'

# 2. Ensure balancer is running
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.setBalancerState(true)'

# 3. Check chunk distribution
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'db.getSiblingDB("config").chunks.count()'

# 4. Add 2 new shards
kubectl patch mongodbsharded my-cluster --type='merge' \
  -p '{"spec":{"shards":{"count":5}}}'

# 5. Monitor balancer progress
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.isBalancerRunning()'

# 6. Verify data distribution
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'printjson(sh.status().shards)'

# 7. Update mongos HPA if needed
kubectl patch hpa my-cluster-mongos -n database --type='merge' \
  -p '{"spec":{"maxReplicas":10}}'
```

## Troubleshooting Scaling Issues

### Scaling Stuck

```bash
# Check operator logs
kubectl logs -n mongodb-operator-system -l app.kubernetes.io/name=mongodb-operator

# Check MongoDB cluster status
kubectl get mongodbsharded my-cluster -o yaml | grep -A 30 status:

# Check pod events
kubectl describe pod my-cluster-shard-3-0 -n database
```

### Balancer Issues

```bash
# Check balancer state
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.getBalancerState()'

# Check forBalancerErrors
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'db.getSiblingDB("config").changelog.find({"what":"balancer.progress"}).sort({time: -1}).limit(5).toArray()'

# Reset balancer if stuck
kubectl exec -it my-cluster-mongos-0 -c mongos -- \
  mongosh -u admin -p $PASSWORD --eval 'sh.stopBalancer(); sh.startBalancer()'
```
