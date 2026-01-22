# E2E 테스트 가이드

## 개요

이 디렉토리는 MongoDB Operator의 End-to-End 테스트를 포함합니다. 이 테스트들은 실제 Kubernetes 클러스터에서 오퍼레이터의 모든 기능을 검증합니다.

## 전제 조건

### 필수 도구

```bash
# Kubernetes 클러스터 (Kind, Minikube, 또는 실제 클러스터)
kind version
# 또는
minikube version

# kubectl
kubectl version --client

# Helm
helm version

# jq (JSON 처리용)
jq --version
```

### Kind 클러스터 생성 (권장)

```bash
# Kind 클러스터 생성 (테스트용)
kind create cluster --name mongodb-test --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
EOF

# 클러스터 확인
kubectl cluster-info --context kind-mongodb-test
kubectl get nodes
```

## 테스트 구조

```
test/e2e/
├── README.md                           # 이 파일
├── run-all-tests.sh                    # 모든 테스트 실행
├── 01-operator-install.sh              # 오퍼레이터 설치 테스트
├── 02-replicaset-basic.sh              # ReplicaSet 기본 테스트
├── 03-replicaset-scale.sh              # ReplicaSet 스케일링 테스트
├── 04-sharded-basic.sh                 # Sharded Cluster 기본 테스트
├── 05-sharded-scale.sh                 # Sharded Cluster 스케일링 테스트
├── 06-backup-restore.sh                # 백업/복원 테스트
├── 07-tls-auth.sh                      # TLS/인증 테스트
├── 08-monitoring.sh                    # 모니터링 테스트
├── 09-failure-scenarios.sh             # 장애 시나리오 테스트
├── 10-cleanup.sh                       # 정리 스크립트
├── manifests/                          # 테스트 YAML 파일
│   ├── replicaset-basic.yaml
│   ├── replicaset-scale.yaml
│   ├── sharded-basic.yaml
│   ├── sharded-scale.yaml
│   ├── backup-s3.yaml
│   ├── backup-pvc.yaml
│   ├── tls-issuer.yaml
│   ├── tls-mongodb.yaml
│   ├── monitoring-mongodb.yaml
│   └── prometheus-stack.yaml
└── utils/                              # 유틸리티 스크립트
    ├── common.sh                       # 공통 함수
    ├── wait-for-ready.sh               # Pod 대기 함수
    ├── mongodb-exec.sh                 # MongoDB 명령 실행
    └── cleanup.sh                      # 리소스 정리
```

## 빠른 시작

### 전체 테스트 스위트 실행

```bash
# 모든 테스트 실행 (약 30-45분 소요)
./test/e2e/run-all-tests.sh
```

### 개별 테스트 실행

```bash
# 1. 오퍼레이터 설치
./test/e2e/01-operator-install.sh

# 2. ReplicaSet 기본 테스트
./test/e2e/02-replicaset-basic.sh

# 3. ReplicaSet 스케일링 테스트
./test/e2e/03-replicaset-scale.sh

# 4. Sharded Cluster 테스트
./test/e2e/04-sharded-basic.sh

# 5. 백업/복원 테스트
./test/e2e/06-backup-restore.sh

# 6. 정리
./test/e2e/10-cleanup.sh
```

## 테스트 세부 정보

### 01. 오퍼레이터 설치 테스트

**목적:** Helm을 통한 오퍼레이터 설치 및 CRD 검증

**테스트 내용:**
- Helm 차트 설치
- CRD 생성 확인
- 오퍼레이터 Pod 실행 확인
- Webhook 준비 상태 확인

**예상 시간:** 2-3분

### 02. ReplicaSet 기본 테스트

**목적:** MongoDB ReplicaSet 배포 및 초기화 검증

**테스트 내용:**
- 3멤버 ReplicaSet 배포
- Pod 준비 상태 대기
- ReplicaSet 초기화 확인
- Primary 선출 확인
- 데이터 삽입 및 조회

**예상 시간:** 5-7분

### 03. ReplicaSet 스케일링 테스트

**목적:** ReplicaSet 수평 확장 검증

**테스트 내용:**
- 초기 3멤버 배포
- 데이터 1000개 삽입
- 5멤버로 스케일 아웃
- 새 멤버 동기화 확인
- 데이터 무결성 검증
- 복제 지연 측정

**예상 시간:** 7-10분

### 04. Sharded Cluster 기본 테스트

**목적:** MongoDB Sharded Cluster 배포 및 샤딩 검증

**테스트 내용:**
- Config Server, Shard, Mongos 배포
- 샤딩 활성화
- 데이터 삽입
- 청크 분산 확인

**예상 시간:** 10-15분

### 05. Sharded Cluster 스케일링 테스트

**목적:** Sharded Cluster 확장 검증

**테스트 내용:**
- 2샤드 → 4샤드 확장
- Mongos 2 → 6 확장
- 청크 밸런싱 확인
- 데이터 분산 검증

**예상 시간:** 15-20분

### 06. 백업/복원 테스트

**목적:** 백업 및 복원 기능 검증

**테스트 내용:**
- PVC 백업 생성
- S3 백업 생성 (S3 자격증명 있는 경우)
- 백업 파일 존재 확인
- 새 클러스터로 복원
- 데이터 무결성 검증

**예상 시간:** 10-15분

### 07. TLS/인증 테스트

**목적:** TLS 암호화 및 인증 검증

**테스트 내용:**
- cert-manager 설치
- Self-signed Issuer 생성
- TLS 활성화된 MongoDB 배포
- 인증서 생성 확인
- TLS 연결 테스트

**예상 시간:** 5-7분

### 08. 모니터링 테스트

**목적:** Prometheus 통합 검증

**테스트 내용:**
- Prometheus Operator 설치
- 모니터링 활성화된 MongoDB 배포
- ServiceMonitor 생성 확인
- 메트릭 수집 확인
- PrometheusRules 검증

**예상 시간:** 5-7분

### 09. 장애 시나리오 테스트

**목적:** 장애 복구 메커니즘 검증

**테스트 내용:**
- Primary 파드 강제 삭제
- 페일오버 시간 측정
- 데이터 손실 여부 확인
- 네트워크 파티션 시뮬레이션
- 자동 복구 확인

**예상 시간:** 10-15분

### 10. 정리

**목적:** 테스트 리소스 정리

**테스트 내용:**
- MongoDB 리소스 삭제
- PVC 삭제
- Namespace 정리
- 오퍼레이터 제거

**예상 시간:** 2-3분

## 환경 변수

테스트는 다음 환경 변수를 지원합니다:

```bash
# 테스트 네임스페이스
export E2E_NAMESPACE="mongodb-e2e"

# 오퍼레이터 네임스페이스
export OPERATOR_NAMESPACE="mongodb-operator-system"

# Helm 차트 경로
export CHART_PATH="./charts/mongodb-operator"

# 타임아웃 (초)
export TIMEOUT=600

# S3 백업 설정 (선택사항)
export S3_BUCKET="mongodb-test-backups"
export S3_REGION="us-east-1"
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"

# 로그 레벨
export LOG_LEVEL="info"  # debug, info, warn, error
```

## 결과 검증

각 테스트는 다음 형식으로 결과를 출력합니다:

```
[PASS] 테스트 이름 - 설명
[FAIL] 테스트 이름 - 오류 메시지
[SKIP] 테스트 이름 - 스킵 사유
```

**전체 요약:**

```
=====================================
E2E Test Results
=====================================
Total Tests:  15
Passed:       13
Failed:       1
Skipped:      1
Duration:     32m 15s
=====================================
```

## 트러블슈팅

### 문제 1: Pod가 Pending 상태

**원인:** 리소스 부족 또는 PV 프로비저닝 실패

**해결:**
```bash
# Pod 상태 확인
kubectl describe pod <pod-name> -n mongodb-e2e

# 노드 리소스 확인
kubectl top nodes

# PV 확인
kubectl get pv
```

### 문제 2: ReplicaSet 초기화 실패

**원인:** 네트워크 문제 또는 MongoDB 버전 불일치

**해결:**
```bash
# MongoDB 로그 확인
kubectl logs <pod-name> -n mongodb-e2e

# 연결 테스트
kubectl exec -it <pod-name> -n mongodb-e2e -- mongosh --eval "rs.status()"
```

### 문제 3: 타임아웃 오류

**원인:** 클러스터 성능 문제

**해결:**
```bash
# 타임아웃 증가
export TIMEOUT=900  # 15분

# 재실행
./test/e2e/02-replicaset-basic.sh
```

## CI/CD 통합

### GitHub Actions

```yaml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Create Kind Cluster
        uses: helm/kind-action@v1
        with:
          cluster_name: mongodb-test

      - name: Install Dependencies
        run: |
          curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
          chmod +x kubectl
          sudo mv kubectl /usr/local/bin/

      - name: Run E2E Tests
        run: ./test/e2e/run-all-tests.sh
```

## 성능 기준

| 테스트 | 예상 시간 | 최대 시간 |
|--------|----------|----------|
| 오퍼레이터 설치 | 2분 | 5분 |
| ReplicaSet 배포 | 5분 | 10분 |
| Sharded Cluster 배포 | 10분 | 20분 |
| 스케일 아웃 (3→5) | 3분 | 7분 |
| 백업 생성 | 2분 | 5분 |
| 복원 | 5분 | 10분 |
| 페일오버 | 30초 | 2분 |

## 기여 가이드

새로운 테스트 추가 시:

1. `test/e2e/XX-test-name.sh` 스크립트 생성
2. `utils/common.sh` 함수 사용
3. 테스트 manifest를 `manifests/` 디렉토리에 추가
4. README.md에 테스트 설명 추가
5. `run-all-tests.sh`에 테스트 추가

**테스트 템플릿:**

```bash
#!/usr/bin/env bash

set -euo pipefail

# 공통 함수 로드
source "$(dirname "$0")/utils/common.sh"

# 테스트 정보
TEST_NAME="Test Name"
TEST_NAMESPACE="${E2E_NAMESPACE:-mongodb-e2e}"

log_info "Starting $TEST_NAME"

# 테스트 로직
# ...

log_success "$TEST_NAME completed"
```

## 참고 자료

- [Kubernetes E2E Testing](https://kubernetes.io/docs/tasks/debug-application-cluster/)
- [Kind 문서](https://kind.sigs.k8s.io/)
- [Helm Testing](https://helm.sh/docs/topics/chart_tests/)
- [MongoDB Operator 문서](../../README.md)
