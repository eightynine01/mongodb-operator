#!/usr/bin/env bash

set -euo pipefail

# 스크립트 디렉토리
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 공통 함수 로드
source "$SCRIPT_DIR/utils/common.sh"

log_info "Starting MongoDB Operator E2E Test Suite"
log_info "========================================"

# 테스트 목록
TESTS=(
    "01-operator-install.sh"
    "02-replicaset-basic.sh"
    "03-replicaset-scale.sh"
    # "04-sharded-basic.sh"
    # "05-sharded-scale.sh"
    # "06-backup-restore.sh"
    # "07-tls-auth.sh"
    # "08-monitoring.sh"
    # "09-failure-scenarios.sh"
)

# 테스트 실행
for test in "${TESTS[@]}"; do
    log_info "Running $test"
    
    if [ -f "$SCRIPT_DIR/$test" ]; then
        "$SCRIPT_DIR/$test" || log_error "$test failed"
    else
        log_warn "$test not found - skipping"
    fi
    
    echo ""
done

log_success "E2E Test Suite Completed"
