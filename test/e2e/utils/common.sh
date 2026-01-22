#!/usr/bin/env bash

# E2E 테스트 공통 함수

# 색상 코드
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로깅 함수
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $*"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

# Pod 준비 대기 함수
wait_for_pods_ready() {
    local namespace=$1
    local label=$2
    local timeout=${3:-600}
    local count=${4:-1}

    log_info "Waiting for $count pod(s) with label '$label' to be ready"

    kubectl wait --for=condition=ready pod \
        -l "$label" \
        -n "$namespace" \
        --timeout="${timeout}s"
}

# MongoDB 명령 실행 함수
mongodb_exec() {
    local namespace=$1
    local pod=$2
    local command=$3

    kubectl exec -it "$pod" -n "$namespace" -- \
        mongosh --quiet --eval "$command"
}

# 환경 변수 초기화
export E2E_NAMESPACE="${E2E_NAMESPACE:-mongodb-e2e}"
export OPERATOR_NAMESPACE="${OPERATOR_NAMESPACE:-mongodb-operator-system}"
export TIMEOUT="${TIMEOUT:-600}"
