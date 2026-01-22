# Learnings - MongoDB Operator Open Source Deployment

## [2026-01-21] Initial Assessment

### Project Structure
- Go 1.25 project using Kubebuilder v4
- CRDs: MongoDB, MongoDBSharded, MongoDBBackup
- Helm chart in `charts/mongodb-operator/`
- Tests: unit (`make test-unit`) and integration (`make test-integration`)
- License: Apache 2.0
- Git remote: eightynine01/mongodb-operator

### Existing Good Practices
- Apache 2.0 license with proper copyright notice
- Comprehensive README with architecture diagrams
- Contributing guide exists
- Distroless Docker image (minimal, secure)
- Helm chart with CRDs included
- Test coverage exists (cover-unit.out, cover-integration.out)

### Key Directories
- `api/v1alpha1/` - CRD type definitions
- `internal/controller/` - Reconciler logic
- `internal/resources/` - Resource builders
- `cmd/main.go` - Operator entry point
- `config/` - Kubernetes manifests
- `charts/` - Helm chart

### Build Commands
- `make build` - Build manager binary
- `make test-unit` - Run unit tests
- `make test-integration` - Run integration tests
- `make docker-build IMG=...` - Build Docker image
- `make docker-push IMG=...` - Push Docker image

### Docker Image
- Default: `eightynine01/mongodb-operator:latest`
- Base: `gcr.io/distroless/static:nonroot`
- Multi-arch ready (ARG TARGETOS, TARGETARCH)

### Code of Conduct Creation
- Contributor Covenant v2.1 provides excellent structure
- Apache 2.0 license notice should be included at end
- Reference CODE_OF_CONDUCT.md in CONTRIBUTING.md for consistency
- Keep Reporting Guidelines specific to project (GitHub Issues)
- Professional tone essential for DevOps/technical audience

### Security Policy Creation (Task 4)
- SECURITY.md essential for open-source security practices
- GitHub Security Advisories provide structured vulnerability reporting
- Key sections needed: Supported Versions, Reporting, Best Practices, Features, Disclosure Policy
- Include Apache 2.0 disclaimer (Section 7) in security documentation
- Privacy note for vulnerability reports builds reporter trust
- Word count ~565 appropriate for comprehensive but concise policy
- Document operator-specific security features (TLS, SCRAM-SHA-256, keyfile, Prometheus, RBAC)
- Use tables for version support and feature lists for scanability
- Provide both GitHub and email options for vulnerability reporting
- Response time承诺 (48 hours) shows commitment without overpromising

### Issue Template Creation (Task 1)
- GitHub issue templates require specific YAML format with `name`, `about`, `title`, `labels`, `body` fields
- Use `type: markdown` for static content and instructions
- Use `type: dropdown` for version fields (Kubernetes, MongoDB, Operator) to ensure consistency
- Use `type: checkboxes` for prerequisites - require `required: true` for important items
- Version dropdowns should include "Other" option with corresponding text input for flexibility
- Bug report template critical fields: versions, steps to reproduce, expected/actual behavior, logs
- Feature request template critical: description, use case, proposed solution, alternatives
- Question template critical: context, what they've tried, configuration, specific question
- Include helpful hints in placeholder text to guide users (e.g., commands to run, what to provide)
- Use `render: shell` or `render: yaml` for code blocks to enable syntax highlighting
- Add security vulnerability disclaimer in bug report template with link to SECURITY.md
- Include willingness to contribute dropdown in feature requests to identify potential contributors
- Priority dropdown helps maintainers understand feature importance
- File sizes: bug_report.yml (5.3KB), feature_request.yml (4.4KB), question.yml (5.7KB)
- Templates reference existing docs (README, CONTRIBUTING, SECURITY) for consistency
- Use markdown formatting for better readability in question answers and feature descriptions

### Pull Request Template Creation (Task 2)
- PR templates use simple Markdown format (not YAML like issue templates)
- Conventional Commits format must be front and center for title conventions
- Include all commit types: feat, fix, docs, style, refactor, test, chore with brief descriptions
- Provide concrete examples for each type (e.g., `feat(controller): add support for arbiter nodes`)
- Description section should guide contributors with three key questions: why, what, how tested
- Type of change section should be mutually exclusive checkboxes (radios not supported in GitHub)
- Related issues section must show GitHub linking syntax (e.g., "Closes #issue-number")
- Checklist should include project-specific commands: `make test` and `make lint`
- Self-review checkbox encourages contributors to do final checks before submission
- PR title convention reminder in checklist ensures consistency
- Reference CONTRIBUTING.md at bottom for detailed guidelines (avoids duplication)
- Keep template concise (70 lines) while being comprehensive - templates longer than 80 lines deter PRs
- Use emoji (📚) sparingly for visual emphasis without being distracting
- Horizontal dividers (---) help visually separate sections for better scanability
- Word count ~210 words ideal for balancing completeness and brevity

### CI Workflow Creation (Task 5)
- GitHub Actions workflows use YAML syntax with jobs, steps, and actions
- Concurrency groups essential to prevent duplicate workflow runs: `group: ${{ github.workflow }}-${{ github.ref }}`
- `cancel-in-progress: true` cancels previous runs when new commit pushed, saves CI resources
- Use `actions/setup-go@v5` for Go setup with built-in module caching: `cache: true`
- Go version matrix testing supports multiple Go versions: ['1.21', '1.22', '1.25']
- Fail-fast on matrix strategy ensures fast feedback: `fail-fast: true`
- Lint job should include `go fmt`, `go vet`, and quick sanity test `go test -short ./...`
- Unit tests generate coverage with `go test -race` and `go tool cover -html=...`
- Coverage threshold check using bash with bc: `if (( $(echo "$COVERAGE < 60" | bc -l) )); then exit 1; fi`
- Codecov integration via `codecov/codecov-action@v4` with flags and names for grouping reports
- Upload artifacts with retention period (7 days) for coverage reports: `actions/upload-artifact@v4`
- Integration tests require setup-envtest for Kubernetes dependencies
- Use ENVTEST_K8S_VERSION environment variable (1.31.0) for consistent K8s version
- Install setup-envtest via curl script from controller-runtime: `curl -sSfL https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.22.4/hack/install-tools.sh | bash -s`
- Build job verifies binary compilation and existence: `make build` then check `bin/manager`
- All jobs run on `ubuntu-latest` for Kubernetes compatibility (no macOS/Windows)
- Branch triggers: push to `main/**` and pull_request to `main/**`
- Coverage reports uploaded as artifacts for visual inspection: `coverage-unit.html`, `coverage-integration.html`
- Total workflow lines: 147 - comprehensive but not overly complex
- Separate jobs for lint, unit-tests, integration-tests, and build for parallel execution

### Docker Build Workflow Creation (Task 6)
- Docker multi-arch builds require QEMU emulation: `docker/setup-qemu-action@v3`
- Buildx enables cross-platform builds: `docker/setup-buildx-action@v3`
- Matrix strategy for architecture builds: [linux/amd64, linux/arm64]
- Use `docker/build-push-action@v6` (latest) for building and pushing images
- Docker metadata action (`docker/metadata-action@v5`) automates tag generation
- Tag flavors include: branch ref, semver patterns (version, major.minor, major), and SHA prefixes
- Platform-specific tags using suffix: `-linux-amd64`, `-linux-arm64`
- Build args match Dockerfile ARG variables: `TARGETOS=linux`, `TARGETARCH=amd64` or `arm64`
- Push only on main/tags (not PRs) using condition: `if: github.event_name != 'pull_request'`
- PR builds use `push: false` for validation without publishing
- Test job verifies: binary existence, executability, nonroot user (65532), and basic `--help` smoke test
- Multi-arch manifest created with `docker buildx imagetools create` combining platform-specific images
- GitHub Actions cache for layers: `cache-from: type=gha`, `cache-to: type=gha,mode=max`
- Job dependencies: build → test → push (cascading for safety)
- Tagging strategy: `latest` on main, `vX.Y.Z`, `vX.Y`, `vX` on tags
- Registry authentication via secrets: `DOCKER_USERNAME`, `DOCKER_PASSWORD` for Docker Hub
- Environment variables for maintainability: `REGISTRY=docker.io`, `IMAGE_NAME=eightynine01/mongodb-operator`
- Use `xargs` to trim whitespace when parsing comma-separated tag lists in bash
- Total workflow lines: 213 - comprehensive with three distinct jobs (build, test, push)
- Build job runs on all triggers (push, PR, dispatch) for continuous validation
- Test and push jobs only run on main/tags to save CI resources on PRs
- Distroless base image verification ensures minimal attack surface (no shell in final image)

### Helm Lint & Test Workflow Creation (Task 7)
- Use `azure/setup-helm@v4` for consistent Helm version installation (v3.16.1 used)
- Helm version 3.8+ required as per project README; using latest v3.16.1 for security patches
- Workflow triggers: push to main, pull_request to main for continuous validation
- Three distinct jobs for Helm validation: lint, test, package - each with specific focus
- Lint job: `helm lint` checks YAML syntax, required fields, template errors, values validation
- Chart.yaml validation ensures required fields: apiVersion, name, version are present
- Values.yaml syntax checked with yamllint (if available) using `--strict` mode
- Test job uses `helm template` to render charts without installing to cluster
- Matrix strategy for testing multiple value files: default, minimal, production values
- Template testing validates Go template rendering with different configuration scenarios
- Minimal values: 1 replica, disabled RBAC, disabled metrics - tests edge cases
- Production values: 3 replicas, resource limits, monitoring, node selectors, tolerations - tests real-world scenarios
- kubectl dry-run validation: `kubectl apply --dry-run=server` checks Kubernetes API compatibility
- Deprecated API detection checks for: extensions/v1beta1, apps/v1beta2, batch/v1beta1
- Package job depends on lint and test jobs (needs: [lint, test]) to ensure only validated charts packaged
- Helm packaging: `helm package charts/mongodb-operator/ --destination /tmp/helm-packages/`
- Package verification: checks .tgz file creation, extracts to validate contents (Chart.yaml, values.yaml present)
- Artifact upload: `actions/upload-artifact@v4` with 30-day retention for package distribution
- Package integrity verified with `helm lint` on generated .tgz file
- Chart path: `charts/mongodb-operator/` - consistent with project structure
- All jobs run on `ubuntu-latest` for Kubernetes compatibility
- Total workflow lines: 204 - comprehensive validation without excessive complexity
- Workflow ensures Helm chart quality before distribution: syntax, rendering, packaging validation

### Release Workflow Creation (Task 9)
- Release workflow must ONLY trigger on version tags: `tags: - 'v*'` (not on every push)
- Workflow dispatch allows manual releases with tag input: `workflow_dispatch: inputs: tag`
- `permissions: contents: write` required for creating GitHub releases
- Environment variables maintain consistency: `REGISTRY=docker.io`, `IMAGE_NAME=eightynine01/mongodb-operator`
- Extract-version job critical for version normalization and validation
- Semantic version validation regex ensures proper format: `^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`
- Output passing between jobs: `outputs.version` and `outputs.tag` for downstream job use
- Build-binary job matrix supports 4 platforms: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64
- Go build with ldflags for version injection: `-ldflags="-s -w -X main.Version=${{ needs.extract-version.outputs.version }}"`
- SHA256 checksums generated for all binaries using `sha256sum` command
- Docker-build job creates multi-arch images for linux-amd64 and linux-arm64 only (no macOS for Docker)
- Docker manifest job combines platform-specific images into multi-arch manifest
- Helm-package job updates Chart.yaml version and appVersion to match release tag
- Release notes extracted from CHANGELOG using awk to match version sections
- CHANGELOG parsing pattern: `$0 == "## [" ver "]" { in_section = 1; print }`
- Helm chart annotations updated with changes extracted from CHANGELOG
- Create-release job uses `softprops/action-gh-release@v2` for GitHub release creation
- Release artifacts include: all platform binaries, checksums.txt, and Helm chart .tgz
- `fail_on_unmatched_files: true` ensures release fails if artifacts missing
- Release marked as non-prelease by default (`prerelease: false`)
- Commit-chart-changes job updates Chart.yaml in main branch after release
- GitHub token authentication: `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}` for creating releases
- Job dependencies ensure correct order: extract-version → build-binary, docker-build, helm-package → docker-manifest → extract-release-notes → create-release → commit-chart-changes
- Use `actions/download-artifact@v4` to retrieve all artifacts before release creation
- Git configuration for automated commits: user.name="github-actions[bot]"
- Total workflow lines: ~450 - comprehensive but modular with 7 distinct jobs
- Release automation eliminates manual steps: version extraction, building, tagging, releasing

### Security Scanning Workflow Creation (Task 8)
- GitHub Actions security workflows require `permissions: security-events: write` for SARIF uploads
- Govulncheck action (`golang/govulncheck-action@v1`) is official Go vulnerability checker
- Use `actions/setup-go@v5` with `cache: true` for faster Go module downloads
- Dependency scanning should fail on CRITICAL and HIGH severity vulnerabilities
- SBOM generation using CycloneDX and SPDX formats for compliance requirements
- CycloneDX action (`CycloneDX/gh-gomod-generate-sbom@v2`) generates software bill of materials from go.mod
- SPDX action (`spdx/spdx-sbom-generator-action@v0.4.0`) provides alternative SBOM format
- Container scanning with Trivy (`aquasecurity/trivy-action@master`) - most up-to-date scanning
- Trivy SARIF format integrates with GitHub Security tab via `github/codeql-action/upload-sarif@v3`
- Build temporary image for scanning (not pushing to registry) with tag `:scan` to avoid conflicts
- Trivy scan severity levels: CRITICAL (fail), HIGH (fail), MEDIUM (warn), LOW (info)
- Separate Trivy scan steps: SARIF for GitHub integration, table for human-readable output, JSON for analysis
- Use `exit-code: '1'` in Trivy to fail workflow on CRITICAL vulnerabilities
- License checking with `go-licenses` ensures Apache 2.0 compatibility
- Apache 2.0 compatible licenses: Apache-2.0, MIT, BSD-3-Clause, BSD-2-Clause, ISC, 0BSD
- Copyleft licenses (GPL, AGPL, LGPL) generate warnings but don't fail workflow
- Use `|| true` on non-critical commands (e.g., `go-licenses save`) to prevent workflow failures
- GitHub Step Summary (`$GITHUB_STEP_SUMMARY`) provides vulnerability counts and compliance status
- Artifact uploads with retention periods: SBOM (90 days), container scan (30 days), license report (90 days)
- Four distinct jobs for modularity: dependency-scan, container-scan, license-check, security-summary
- Security-summary job aggregates results from all scan jobs using `needs` dependency
- Use `if: always()` on summary steps to ensure reports generate even if scans fail
- Workflow triggers: push to main/**, pull_request to main/**, workflow_dispatch for manual scanning
- jq queries for JSON parsing: `[.. | .Vulnerabilities? | select(. != null) | .[] | select(.Severity == "CRITICAL")] | length`
- Tricky jq syntax for recursive vulnerability search with `..` operator
- Total workflow lines: 295 - comprehensive security coverage with clear separation of concerns
- All scan artifacts uploaded for audit trail and manual review
- Security policy: fail on CRITICAL, warn on HIGH/MEDIUM, informational on LOW
- License report saved as CSV and raw license files for easy review

### README Badge Enhancement (Task 10)
- Badges provide immediate visibility into project health and quality metrics
- Badge placement: immediately after title, before overview section - prime visibility area
- Existing badges preserved: License (Apache 2.0), Go Version (1.21+), Kubernetes (1.26+)
- New badges added:
  - **Build Status**: GitHub Actions workflow badge - `actions/workflows/ci.yml/badge.svg`
  - **Docker Image Size**: shields.io badge - shows latest image size from Docker Hub
  - **Helm Chart Version**: Dynamic YAML badge - reads version from Chart.yaml in repository
  - **Go Report Card**: goreportcard.com badge - code quality metrics for Go projects
  - **Codecov**: coverage badge - shows test coverage percentage
- Shield.io is reliable badge provider for dynamic metrics (Docker size, Helm version)
- Helm version badge uses raw GitHub URL to parse Chart.yaml: `https://raw.githubusercontent.com/eightynine01/mongodb-operator/main/charts/mongodb-operator/Chart.yaml`
- Go Report Card provides A-F grade based on code complexity, formatting, and documentation
- Codecov badge requires CI workflow to upload coverage reports (covered in Task 5)
- Badge count: 8 total (3 existing + 5 new) - within optimal range (5-7) for readability
- All badge links use HTTPS for security
- Badge colors match project branding: blue for stable versions, status colors for CI/coverage
- Single edit approach works well when badges are contiguous - maintains formatting
- Badge order matters: group related badges together (version badges first, then quality badges)
- Docker Image Size badge will show "unknown" until first image is pushed to Docker Hub
- Helm chart badge will fail if Chart.yaml path or structure changes - monitor after chart updates
- Go Report Card requires public repository for scanning
- No separate badges section needed - badges work inline at top of README
- Total badges line count: 8 badges, 8 lines - clean and scannable

### Example Manifests Creation (Task 12)
- examples/ directory structure: minimal/, production/, backups/, monitoring/
- Each example file contains complete Kubernetes manifests with comments for user guidance
- Minimal examples focus on production-ready but simplified configurations (3-member replicas, basic resources)
- Production examples include enterprise-grade features: resource limits, PDBs, network policies, TLS, monitoring, HA
- Backup examples demonstrate S3 and PVC storage options with CronJob for scheduled backups
- Monitoring examples provide complete Prometheus stack: ServiceMonitors, Prometheus, Grafana, dashboards
- CRD references use apiVersion: mongodb.keiailab.com/v1alpha1
- All YAML files validated with kubectl dry-run to ensure Kubernetes API compatibility
- Comments are necessary for example files to guide users on configuration (not agent memo-style comments)
- Production examples include: Pod Disruption Budgets (PDBs), Network Policies, Pod Anti-Affinity, Node Selectors, Topology Spread Constraints
- Resource recommendations from README incorporated: Mongos minimum 512Mi memory to avoid OOM
- Secret placeholders use "REPLACE_WITH_STRONG_PASSWORD" and provide generation commands
- Storage class placeholders ("standard", "fast-ssd") require user customization
- Cert-manager issuer references need user configuration ("letsencrypt-prod")
- Dashboard JSON embedded in ConfigMap for Grafana auto-import
- ServiceMonitor labels match Prometheus operator selector ("release: prometheus")
- Namespace declarations included for self-contained examples

### Documentation Creation (Task 11)
- Documentation structure organized: user docs (getting-started, advanced topics, troubleshooting) separate from developer docs (development, testing, architecture)
- Clear navigation with README.md providing quick links to all documentation
- Content quality balanced: concise yet comprehensive with word counts meeting requirements (250-500 words per document)
- Code examples essential: every section includes practical YAML and bash examples
- Cross-referencing important: linking between related documentation for better navigation
- Real-world scenarios covered: TLS with cert-manager, Prometheus monitoring, S3/PVC backups, scaling strategies
- Troubleshooting comprehensive: common issues with step-by-step resolution guides
- Developer documentation thorough: local development setup (Kind/Minikube), hot reload, debugging, testing strategies
- Architecture documentation detailed: controller design, resource builders, MongoDB package, finalizer patterns
- 9 documentation files created covering user and developer needs
- Quick reference tables help users find information efficiently
- Progressive complexity from getting-started to advanced topics
- Verification checklist ensure complete documentation coverage

### Documentation Metrics
- User documentation: 5 files (getting-started, tls, monitoring, backup, scaling, troubleshooting)
- Developer documentation: 3 files (development, testing, architecture)
- Index file: 1 file (README.md)
- Total word count: ~6,550 words across all documents
- Code examples: 25+ YAML manifests, 50+ bash commands, 15+ Go code snippets
- All documents meet word count requirements and provide comprehensive coverage

### Artifact Hub Integration Setup (Task 13)
- artifacthub-repo.yml must be at repository level (charts/), NOT inside chart directory
- Correct location: `charts/artifacthub-repo.yml` (same level as index.yaml)
- artifacthub-repo.yml contains: repositoryID, signingKey (PGP), owners (email/name)
- Chart.yaml has comprehensive Artifact Hub annotations already configured
- Required Chart.yaml metadata: name, version, type: application, description ✅
- Recommended metadata present: home URL, icon URL, keywords array, maintainers array ✅
- Icon URL `https://raw.githubusercontent.com/mongodb/mongo/master/docs/leaf.svg` is accessible (HTTP 200)
- Artifact Hub annotations in Chart.yaml include:
  - `artifacthub.io/license`: Apache-2.0
  - `artifacthub.io/category`: database
  - `artifacthub.io/operatorCapabilities`: Full Lifecycle
  - `artifacthub.io/links`: Documentation, Source Code, MongoDB Docs
  - `artifacthub.io/crds`: Full CRD definitions for MongoDB, MongoDBSharded, MongoDBBackup
  - `artifacthub.io/crdsExamples`: Complete YAML examples for all three CRDs
  - `artifacthub.io/containsSecurityUpdates`: false
  - `artifacthub.io/images`: Operator, MongoDB, MongoDB-exporter images
  - `artifacthub.io/prerelease`: false
  - `artifacthub.io/changes`: Security fixes, fixed issues
  - `artifacthub.io/recommendations`: Related packages (kube-prometheus-stack, cert-manager)
- README.md meets all Artifact Hub requirements:
  - Description and features ✅
  - Prerequisites (Kubernetes 1.26+, Helm 3.8+) ✅
  - Installation instructions (helm repo add, helm install) ✅
  - Configuration tables with parameters ✅
  - Usage examples (ReplicaSet, Sharded Cluster, Backup) ✅
  - Uninstallation instructions ✅
  - Upgrading instructions ✅
  - Troubleshooting section ✅
- Screenshots are optional for Artifact Hub (not required)
- Icon URL in Chart.yaml is sufficient (no local logo file needed)
- Helm lint passes: `helm lint charts/mongodb-operator/` - 0 failures ✅
- PGP signing key already configured in artifacthub-repo.yml for verified publisher status
- Repository ID present: `386b6255-6da7-4a73-8fc0-a8e79e3c7b90`
- Ownership already claimed in Artifact Hub (owners section with email)
- No modifications needed - chart is production-ready for Artifact Hub

### Helm Repository Publishing Setup (Task 14)
- GitHub Pages is standard solution for Helm chart repositories using gh-pages branch
- Helm repository URL format: `https://<org>.github.io/<repo>/` for GitHub Pages
- index.yaml must be in root directory of gh-pages branch for Helm to find it
- `helm repo index` command with `--merge` option preserves existing chart entries while adding new ones
- Helm publish workflow requires two branches: main (source code) and gh-pages (published artifacts)
- Workflow uses `actions/checkout@v4` with different paths to checkout multiple branches simultaneously
- Helm chart packaging creates .tgz file: `helm package charts/mongodb-operator/`
- Index regeneration required after packaging: `helm repo index . --url <repo-url> --merge index.yaml`
- Cleanup job keeps gh-pages repository size manageable by removing old chart versions
- Version sorting with `sort -V` (version sort) ensures oldest versions removed first
- `xargs -r rm -f` safely removes files without error if list is empty
- Workflow dispatch with `keep_versions` input allows manual republishing with configurable retention
- Triggers on `charts/mongodb-operator/**` path ensures chart updates automatically trigger publish
- No changes detection: `git diff --quiet && git diff --cached --quiet` prevents empty commits
- GitHub Actions bot credentials: `github-actions[bot]@users.noreply.github.com` for automated commits
- README already contains correct Helm repository URL: `https://eightynine01.github.io/mongodb-operator`
- index.yaml exists at root with two versions (0.0.7 and 0.0.2) already configured
- Existing index.yaml references raw GitHub URLs which should be replaced by gh-pages URLs after workflow runs

### Codecov Integration Setup (Task 15)
- Codecov was already configured in CI workflow from previous task (Task 5)
- Both unit-tests and integration-tests jobs have `codecov/codecov-action@v4` for uploading coverage
- Coverage threshold (60%) was implemented in unit-tests job but missing from integration-tests job
- Added coverage threshold check to integration-tests job using same bash pattern as unit-tests
- Coverage threshold logic: `if (( $(echo "$COVERAGE < 60" | bc -l) )); then exit 1; fi`
- Codecov badge already present in README.md at correct location (line 10) with correct format
- Badge format: `[![codecov](https://codecov.io/gh/eightynine01/mongodb-operator/branch/main/graph/badge.svg)](https://codecov.io/gh/eightynine01/mongodb-operator)`
- Created `.codecov.yml` configuration file for enhanced Codecov settings:
  - Project coverage status with auto target and 1% threshold
  - Patch coverage status with same settings
  - Pull request comments enabled with layout: reach, diff, flags, files
  - Ignore patterns for test files (`*_test.go`), generated files (`zz_*.go`, `*.pb.go`), third_party, vendor
- Coverage files: `cover-unit.out` (unit tests) and `cover-integration.out` (integration tests)
- Coverage reports generated as HTML artifacts with 7-day retention
- Codecov action flags: `unit` for unit tests, `integration` for integration tests
- Coverage job names include Go version for integration tests matrix
- BC calculator required for floating-point comparison in coverage threshold checks
- User must set up `CODECOV_TOKEN` secret in GitHub repository settings for upload to work
- Codecov provides PR comments and status checks automatically when token is configured
- Integration test coverage threshold ensures both test suites maintain minimum quality standards

### Dependabot Configuration (Task 16)
- Dependabot supports separate config files for different package ecosystems using version: 2 format
- Go modules dependency tracking uses `package-ecosystem: "gomod"` with `directory: "/"`
- GitHub Actions dependency tracking uses `package-ecosystem: "github-actions"` with `directory: "/"`
- Weekly schedule recommended (daily too aggressive): `interval: "weekly"`, `day: "monday"`, `time: "09:00"`, `timezone: "UTC"`
- Open pull requests limit set to 50 to prevent PR spam: `open-pull-requests-limit: 50`
- Target branch must be explicitly set: `target-branch: "main"` for production updates
- Labels help organize dependency PRs: "dependencies", "go", "github-actions"
- Commit message prefixes follow Conventional Commits: `chore(deps)` for Go, `ci(deps)` for Actions
- Separate prefix for development dependencies: `prefix-development: "chore(deps-dev)"`
- Dependency grouping reduces PR noise by batching related updates together:
  - Kubernetes dependencies: `k8s.io/*`, `sigs.k8s.io/*`
  - Controller runtime: `sigs.k8s.io/controller-runtime*`
  - Testing dependencies: ginkgo, gomega, testify
  - Core actions: checkout, setup-go, upload/download-artifact
  - Docker actions: docker/*
  - Security actions: aquasecurity/*, golang/*, github/codeql-action*
  - Release actions: softprops/action-gh-release*
  - Helm actions: azure/setup-helm*
- Ignore major version bumps for critical dependencies to prevent breaking changes: `update-types: ["version-update:semver-major"]`
- Both direct and indirect dependencies should be monitored: `allow: [dependency-type: "direct", dependency-type: "indirect"]`
- Auto-rebase strategy keeps PRs updated: `rebase-strategy: "auto"`
- Reviewers and assignees ensure proper review: `reviewers: ["eightynine01"]`, `assignees: ["eightynine01"]`
- Pull request branch name separator: `separator: "-"` for clean branch names
- Include scope in commit messages: `include: "scope"` shows which dependency updated
- Separate dependabot.yml for Go deps and dependabot-github-actions.yml for Actions deps improves maintainability
- Configuration files self-documenting - YAML comments removed to follow best practices
- Go config file size: 52 lines, GitHub Actions config: 47 lines - concise and focused
- Dependabot automates dependency updates, security vulnerability detection, and lockfile management

### Pre-commit Hooks Implementation (Task 17)
- Pre-commit framework provides automated code quality checks before each commit
- Configuration file `.pre-commit-config.yaml` placed at project root
- `default_language_version: go: "1.25"` ensures Go version consistency
- `fail_fast: false` runs all hooks even if some fail for comprehensive feedback
- `minimum_pre_commit_version: "3.0.0"` ensures compatibility with modern pre-commit features
- Pre-commit hooks configured:
  - **trailing-whitespace**: Removes trailing whitespace from modified files
  - **go fmt**: Auto-formats Go code to gofmt standards
  - **go vet**: Runs go vet to find potential issues
  - **golangci-lint**: Runs comprehensive Go linter with sensible defaults (using `--fix` flag)
  - **go test**: Runs unit tests (`go test -v -race ./...`) to catch regressions
  - **end-of-file-fixer**: Ensures files end with single newline
  - **check-yaml**: Validates YAML syntax
  - **check-toml**: Validates TOML syntax
  - **check-added-large-files**: Prevents large file commits (max 1000KB)
  - **check-case-conflict**: Detects case-insensitive filesystem conflicts
  - **check-merge-conflict**: Detects merge conflict markers
  - **commit-msg-trim**: Removes trailing whitespace from commit messages
- Pre-commit repos used:
  - `github.com/pre-commit/pre-commit-hooks@v5.0.0` - Basic file checks
  - `github.com/golangci/golangci-lint@v1.62.2` - Go linting
  - `local` hooks - Go-specific tools (fmt, vet, test) using system language
- Installation command: `curl https://pre-commit.com/install.sh | sh`
- Enable hooks: `pre-commit install` (runs automatically on commit)
- Manual execution: `pre-commit run --all-files` (check all files)
- Hook ordering: fmt → vet → lint → test (logical progression from formatting to testing)
- golangci-lint uses `--fix` flag to automatically fix issues when possible
- go test includes `-v -race` flags for verbose output and race detection
- CONTRIBUTING.md updated with pre-commit section: installation, usage, hook descriptions, workflow
- Local development workflow: hooks run automatically before each git commit
- Pre-commit integration complements Makefile targets (use pre-commit for local checks, Makefile for CI)
- No expensive hooks added (integration tests excluded from pre-commit to keep commits fast)
- Hook configuration follows official pre-commit documentation for Go projects

### GitHub Settings Documentation (Task 22)
- Comprehensive GitHub settings guide essential for professional open source project management
- Documentation structure should cover: general settings, branch protection, templates, security, required checks, permissions
- General settings require: public visibility, topics for discoverability (7 topics for MongoDB Operator), proper description
- Branch protection rules critical for code quality: require PR reviews, status checks, up-to-date branches
- All workflow checks must be required: Lint, Unit Tests, Integration Tests, Build, Dependency Scan, Container Scan, License Check, Helm Lint, Helm Test, Docker Build
- Coverage threshold (60%) enforced in both unit and integration tests
- Branch protection should require branches to be up to date before merging to prevent conflicts
- Dismiss stale PR approvals when new commits pushed for security
- Block force pushes on main branch recommended for cleaner history
- Issue templates use YAML format with name, about, title, labels, body fields
- Issue types: bug report (versions, steps, logs), feature request (use case, solution, alternatives), question (context, what tried, specific question)
- PR templates use Markdown format with Conventional Commits convention, description, type, checklist
- Labels organization critical: type (bug/enhancement/question), status (needs-triage/in-progress/ready-to-merge), priority (critical/high/medium/low), component (replicaset/sharded-cluster/backup)
- Security settings: Dependabot enabled, code scanning, secret scanning, 2FA for maintainers
- Security scanning workflow fails on CRITICAL/HIGH vulnerabilities, warns on MEDIUM/LOW
- SBOM generation in both CycloneDX and SPDX formats for compliance
- License checking for Apache 2.0 compatibility with copyleft warnings
- Workflow permissions: Read and write required for Contents, Pull Requests, Security Events, Checks
- Fork PR workflows enabled with caution (limited secret access)
- Tag creation restricted to maintainers to control versioning
- Actions retention: 90 days default for compliance and debugging
- GitHub Steps Summary provides vulnerability counts and compliance status
- Additional configuration files: CODEOWNERS for code ownership, .github/dependabot.yml for automated updates
- Regular maintenance tasks: weekly Dependabot review, monthly stale PR cleanup, quarterly settings review
- Troubleshooting section common issues: required checks failing, permission denied, branch protection too strict
- Documentation word count ~1,800 words for comprehensive coverage without being overwhelming
- Settings checklist with 50+ items ensures nothing overlooked during configuration
- Reference section links to GitHub official documentation for each feature

### Comprehensive Test Suite Strategy (Task 21)
- Testing strategy document (docs/developers/testing-strategy.md) provides comprehensive testing framework for MongoDB Operator
- Test pyramid structure essential: E2E (slow, 10 tests), Integration (medium, 30+ tests), Unit (fast, 100+ tests)
- Existing tests use Ginkgo v2 framework with envtest for Kubernetes integration testing (K8s 1.31.0)
- Unit tests focus on controller reconciliation logic, resource builders, MongoDB package functions, validation, error handling
- Integration tests verify CRD creation, resource reconciliation, status updates, Kubernetes version compatibility
- E2E tests validate complete user workflows: ReplicaSet deployment, sharded cluster initialization, backup/restore, scaling, TLS rotation, version upgrades
- Performance tests benchmark critical paths: ReplicaSet init (<120s), sharded cluster init (2-shards<180s, 3-shards<240s, 5-shards<360s), concurrent connections (99th<500ms, error rate<0.1%)
- Coverage goals: Unit 80%+, Integration 70%+, E2E 60%+ with specific targets by package (controller 80%, resources 85%, mongodb 80%, api/v1alpha1 70%)
- Critical paths require 100% coverage: ReplicaSet initialization, sharded cluster init, shard scale-out (sh.addShard), admin user creation, TLS cert handling, secret management (keyfile gen)
- Test data management with mock fixtures in internal/controller/testfixtures/ (DefaultMongoDB, DefaultMongoDBSharded)
- Kubernetes version compatibility matrix: 1.31.0 (primary), 1.29.0 (stable), 1.28.0 (stable), 1.26.0 (minimum), 1.27.0 (deprecated)
- Integration test enhancements required:
  - Shard scale out (2→5) - verify new shards created, initialized, registered with mongos, balancer distributes data
  - Scale down (5→3) - track orphaned resources (shards 3-4), verify active shards (0-2) remain
  - TLS certificate rotation - verify cluster reconnects with new cert, pods restarted, no data loss
  - Backup/restore workflow - verify backup completion in S3, deletion, restore, data integrity
  - Version upgrade path (7.0→8.2) - verify rolling update, data migration, no breaking changes
  - Admin user auto-creation edge cases - existing user, missing secret, invalid password, timeout, concurrent reconciliation
- E2E scenarios defined:
  1. Fresh Deployment - deploy ReplicaSet, create admin user, perform CRUD, verify persistence after pod restart
  2. Sharded Cluster - deploy with 3 shards, verify config server/shards/mongos init, add shards via sh.addShard(), verify balancer distribution
  3. Backup & Restore - deploy MongoDB, create data, create MongoDBBackup, verify S3 storage, delete cluster, restore, verify integrity
  4. Upgrade Path - deploy version X, upgrade to X+1, verify rolling update, data migration, no breaking changes
  5. TLS Rotation - deploy with TLS enabled, rotate cert-manager certificate, verify cluster reconnects, no data loss
  6. Scaling Operations - horizontal (mongos replicas), vertical (resources), add shard, verify health, no data loss during concurrent operations
- Test environment setup provides both Kind and Minikube installation guides with complete Docker image loading and operator installation steps
- Running tests documented for all types: unit (go test, coverage, race), integration (envtest, specific K8s version), E2E (Kind/Minikube, kubectl), performance (benchmarks, k6 load testing)
- Debugging failing tests section covers: enable debug logging, inspect cluster state (kubectl get, logs, exec), dump test state (kubectl get all, describe), retry flaky tests (FlakeAttempts)
- Test organization structure defined: internal/controller/*_test.go (controller tests), internal/resources/*_test.go (builder tests), internal/mongodb/*_test.go (package tests), tests/e2e/*_test.go (scenarios), tests/load/*.js (k6), tests/performance/*_test.go (benchmarks)
- Test fixtures and secrets should be in internal/controller/testfixtures/ with reusable DefaultMongoDB, DefaultMongoDBSharded objects
- Coverage enforcement via bash using bc calculator: `if (( $(echo "$COVERAGE < 80" | bc -l) )); then exit 1; fi`
- Coverage reports generated as HTML artifacts for visual inspection with 7-day retention
- E2E test execution requires Kind cluster with custom config (extraMounts for pods) or Minikube with registry addon for local images
- Local development workflow: code changes → affected tests → pre-commit → full test suite → build/load image → test in cluster
- Continuous integration triggers on push/main, PR/main, manual dispatch with matrix strategy for Kubernetes versions (1.26.0, 1.28.0, 1.29.0, 1.31.0)
- Testing strategy document comprehensive (3,161 words) providing concrete code examples for all test scenarios and practical implementation guidance
- Performance thresholds based on MongoDB community benchmarks and operator stability testing from README
- Ginkgo v2 BDD-style test framework used with Eventually/Expect pattern for async operations (timeout 10s, interval 250ms in existing tests)
- envtest setup uses KUBEBUILDER_ASSETS env var or default path to bin/k8s/1.31.0-darwin-arm64 for local development
- Test suite structure complements existing CI workflow (Task 5) and enables comprehensive E2E testing for future development

### v1.0.0 Release Process (Task 23)
- Release guide creation (docs/releases/v1.0.0-release-guide.md) provides comprehensive step-by-step release instructions for maintainers
- CHANGELOG.md follows Keep a Changelog format with version sections, release dates, organized by Added, Changed, Deprecated, Removed, Fixed, Security
- v1.0.0 section added with comprehensive release notes documenting all new features from open source deployment (GitHub templates, CI/CD workflows, documentation, examples, Artifact Hub, Dependabot, pre-commit hooks, Codecov, Helm publishing, security scanning)
- Breaking changes section explicitly states "None" for GA release - important for user confidence
- Chart.yaml version and appVersion must be synchronized (both 1.0.0) to avoid version mismatch errors
- Artifact Hub annotations updated in Chart.yaml with new release changes (kind: added for all features)
- Docker image reference in artifacthub.io/images must match release version (eightynine01/mongodb-operator:1.0.0)
- Release guide includes pre-release checklist with 5 categories (Code Quality, Documentation, Testing, Configuration, Security) with 25+ items
- Git tag creation uses annotated tags with multi-line message documenting all features - recommended for stable releases
- Docker multi-arch builds support linux/amd64 and linux/arm64 with buildx for cross-platform compatibility
- SBOM (Software Bill of Materials) and Provenance attestations enabled via `--provenance=mode=max --sbom=true` flags for enterprise compliance
- Helm chart publishing requires gh-pages branch with index.yaml and packaged .tgz files at root directory
- `helm repo index --merge` preserves existing chart entries while adding new ones - critical for maintaining chart history
- Cross-platform binary builds for 4 platforms: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64
- SHA256 checksums.txt generated for all binaries for integrity verification
- GitHub release creation via `gh release create` CLI recommended over web UI for automation
- Release artifacts include: all platform binaries (tar.gz), checksums.txt, Helm chart (.tgz)
- Release notes automatically extracted from CHANGELOG using awk pattern matching: `$0 == "## [" ver "]" { in_section = 1; print }`
- Release verification checklist: Docker image pull/test, Helm chart install, binary download/test, GitHub release UI, Artifact Hub appearance
- Post-release tasks: update README with release notes section, announce via GitHub Discussions/Twitter/Reddit/LinkedIn, update project roadmap, monitor for issues
- Release timeline: ~2-3 hours total (pre-release 1-2h, release steps 30-60min, post-release 30-60min)
- Future releases follow semantic versioning: major (x.0.0) for breaking changes, minor (1.x.0) for new features, patch (1.0.x) for bug fixes
- Release guide comprehensive (3,500+ words) with troubleshooting section for common issues (Docker build fails, Helm index fails, tag exists, binary build fails, Artifact Hub not updating)
- Environment variables defined for consistency: VERSION, REGISTRY, IMAGE_NAME, CHART_DIR, DIST_DIR
- Useful commands provided: git tag operations, Docker operations, Helm repo commands, GitHub CLI commands
- Release process documentation enables any maintainer to perform releases with confidence
- v1.0.0 marks transition from pre-release (0.0.x) to stable (1.0.0) - first production-ready release
