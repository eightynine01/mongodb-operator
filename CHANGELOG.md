# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Point-in-Time Recovery (PITR)
- Automated version upgrades
- Cross-cluster replication
- Grafana dashboard templates
- Shard scale-in support

## [0.0.3] - 2024-12-31

### Added
- Automatic ReplicaSet initialization with `rs.initiate()`
- Automatic Sharded Cluster initialization
  - Config server ReplicaSet initialization
  - Shard ReplicaSet initialization
  - Automatic `sh.addShard()` execution
- Admin user auto-creation via MongoDB localhost exception
- Horizontal shard scaling (scale out) support
- Resource recommendations documentation
- Tested features and limitations documentation
- Mongos replica scaling examples

### Fixed
- Preserve shard status arrays during scale out (prevent re-initialization)
- Port configuration: configsvr (27019), shardsvr (27018), mongos (27017)
- Keyfile Secret regeneration causing authentication failures
- Mongos readiness probe timeout (increased to 5s)
- Container-aware command execution for mongos pods

### Changed
- Marked as stable release (prerelease: false)
- Minimum mongos memory recommendation: 512Mi

## [0.0.2] - 2024-12-24

### Added
- Verified Publisher configuration for ArtifactHub

### Changed
- Updated repository metadata with repository ID (386b6255-6da7-4a73-8fc0-a8e79e3c7b90)

## [0.0.1] - 2024-12-23

### Added
- Initial pre-release for testing
- MongoDB ReplicaSet CRD and controller
  - Support for 3+ member replica sets
  - Automatic keyfile generation for internal authentication
  - SCRAM-SHA-256 authentication support
  - Arbiter node support
- MongoDB Sharded Cluster CRD and controller
  - Config server replica set management
  - Multiple shard support with configurable members per shard
  - Mongos router deployment with auto-scaling (HPA)
- MongoDBBackup CRD and controller
  - S3-compatible storage support
  - PVC-based backup storage
  - Full and incremental backup types
  - Compression support (gzip, zstd, snappy)
- TLS encryption support
  - cert-manager integration for automatic certificate management
  - Self-signed certificate option
- Prometheus monitoring integration
  - MongoDB exporter sidecar
  - ServiceMonitor resource creation
  - PrometheusRule for alerts
- Helm chart for easy deployment
  - Configurable values for all operator settings
  - CRD installation via Helm
  - RBAC resources included

### Security
- Non-root container execution
- Read-only root filesystem
- Dropped capabilities
- SeccompProfile enforcement

---

[Unreleased]: https://github.com/eightynine01/mongodb-operator/compare/v0.0.3...HEAD
[0.0.3]: https://github.com/eightynine01/mongodb-operator/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/eightynine01/mongodb-operator/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/eightynine01/mongodb-operator/releases/tag/v0.0.1
