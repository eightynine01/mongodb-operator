# Decisions - MongoDB Operator Open Source Deployment

## [2026-01-21] Technology Choices

### CI/CD Platform
- **Decision**: Use GitHub Actions (native to GitHub)
- **Rationale**: Free for public repos, excellent integration, large community
- **Alternatives considered**: GitLab CI, CircleCI, Travis CI

### Code Coverage Tool
- **Decision**: Codecov
- **Rationale**: Free for open source, excellent GitHub integration, good reports
- **Alternatives considered**: Coveralls, Code Climate

### Dependency Management
- **Decision**: Dependabot (built into GitHub)
- **Rationale**: No additional setup, integrates with PRs, free for open source
- **Alternatives considered**: Renovate, Snyk

### Pre-commit Hooks
- **Decision**: pre-commit framework
- **Rationale**: Language-agnostic, easy to configure, widely used
- **Alternatives considered**: husky (JS-focused), shell scripts

### Release Automation
- **Decision**: GitHub Release workflow + semantic-release style manual
- **Rationale**: Keep it simple, no external dependencies, full control
- **Alternatives considered**: semantic-release, go-release

## [2026-01-21] v1.0.0 Release Decisions

### Release Version
- **Decision**: v1.0.0 as initial stable GA release
- **Rationale**: Comprehensive CI/CD, documentation, and examples ready. All 23 tasks from open source deployment plan complete. Breaking changes: none.
- **Alternatives considered**: v0.1.0 (continue pre-release), v1.0.0-beta.1 (beta release)

### Release Documentation
- **Decision**: Comprehensive release guide in docs/releases/v1.0.0-release-guide.md
- **Rationale**: Guides maintainers through complete release process with troubleshooting. Enables future releases with confidence.
- **Alternatives considered**: Minimal README instructions, external wiki

### CHANGELOG Format
- **Decision**: Keep a Changelog format with sections for Added, Changed, Breaking Changes, Security
- **Rationale**: Industry standard, user-friendly, machine-parsable for release automation
- **Alternatives considered**: Simple bullet points, GitHub Releases only, NEWS file

### Docker Image Strategy
- **Decision**: Multi-arch builds (linux/amd64, linux/arm64) with SBOM and Provenance
- **Rationale**: Modern cloud requirements (ARM64 servers), enterprise compliance (SBOM), build integrity (Provenance)
- **Alternatives considered**: Single architecture (amd64 only), no SBOM/Provenance

### Helm Chart Publishing
- **Decision**: GitHub Pages (gh-pages branch) for Helm repository
- **Rationale**: Free, simple, integrates with GitHub Actions, no external infrastructure
- **Alternatives considered**: Cloud Storage (S3), ChartMuseum, Harbor

### Binary Distribution
- **Decision**: Cross-platform binaries (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64) with checksums
- **Rationale**: User flexibility for local development and testing, integrity verification via SHA256
- **Alternatives considered**: Docker images only, single architecture binaries

### Release Verification
- **Decision**: Comprehensive 5-step verification (Docker, Helm, binaries, GitHub, Artifact Hub)
- **Rationale**: Prevents broken releases, ensures all artifacts working, builds user trust
- **Alternatives considered**: Minimal verification, automated only, user testing

### Post-Release Announcements
- **Decision**: Multi-channel announcements (GitHub Discussions, Twitter, Reddit, LinkedIn)
- **Rationale**: Reaches broad audience, builds community, drives adoption
- **Alternatives considered**: GitHub only, email newsletter, no announcement

### Semantic Versioning for Future
- **Decision**: Follow semantic versioning strictly (MAJOR.MINOR.PATCH)
- **Rationale**: Predictable releases, user confidence, industry standard
- **Alternatives considered**: Calendar versioning (CalVer), continuous versioning, no strict versioning

### Release Timeline
- **Decision**: ~2-3 hours total release time (pre-release 1-2h, release 30-60min, post-release 30-60min)
- **Rationale**: Balances thoroughness with efficiency. Pre-release testing prevents re-releases.
- **Alternatives considered**: Minimal 30-minute release, full day release testing

### Release Guide Content
- **Decision**: 9-section guide (Pre-release, Process, Troubleshooting, Future Releases, Appendix)
- **Rationale**: Comprehensive without overwhelming. Troubleshooting section prevents stuck releases.
- **Alternatives considered**: Minimal checklist, interactive wizard, video tutorial

### Automated Release vs Manual
- **Decision**: Semi-automated (GitHub Actions builds, manual release creation)
- **Rationale**: Balance automation with human oversight. Prevents accidental releases.
- **Alternatives considered**: Fully automated (auto-release on tag), fully manual (everything manual)

### Tag Push Timing
- **Decision**: Push tag during GitHub release creation (not before)
- **Rationale**: Ensures release artifacts ready before tag goes public. Prevents premature downloads.
- **Alternatives considered**: Push tag immediately, push after release

### Artifact Hub Updates
- **Decision**: Rely on automatic Artifact Hub sync (no manual trigger)
- **Rationale**: Artifact Hub monitors GitHub releases automatically. ~24 hour delay acceptable.
- **Alternatives considered**: Manual trigger, webhook integration, no Artifact Hub

### Release Checklist Items
- **Decision**: 25+ pre-release checklist items across 5 categories
- **Rationale**: Comprehensive catch for common issues. Reduces post-release bug reports.
- **Alternatives considered**: Minimal 5-item checklist, no checklist, fully automated checks only

### Helm Version Synchronization
- **Decision**: Chart.yaml version and appVersion must match (both 1.0.0)
- **Rationale**: Prevents user confusion. Consistent versioning across all artifacts.
- **Alternatives considered**: Separate chart/app versions, appVersion tracks operator version only
