#!/usr/bin/env bash
set -euo pipefail

# Local CI checks for the Sigil CLI (Go module at repo root).
# Usage: ./scripts/test-ci.sh

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; TESTS_PASSED=$((TESTS_PASSED + 1)); }
log_error()   { echo -e "${RED}[✗]${NC} $1";   TESTS_FAILED=$((TESTS_FAILED + 1)); }
log_warning() { echo -e "${YELLOW}[!]${NC} $1"; }

run_cmd() {
    local name="$1"
    shift
    log_info "Running: $name"
    if (eval "$@"); then
        log_success "$name"
    else
        log_error "$name"
        FAILED_TESTS+=("$name")
    fi
}

echo "=========================================="
echo "  Sigil CLI — local CI"
echo "=========================================="
echo ""

log_info "=== Prerequisites ==="
if ! command -v go &>/dev/null; then
    log_error "go not found in PATH"
    exit 1
fi
log_info "Go: $(go version)"
if ! command -v gpg &>/dev/null && ! command -v gpg2 &>/dev/null; then
    log_warning "gpg/gpg2 not in PATH — internal/gpg tests will be skipped"
else
    log_info "GnuPG available (integration tests may run)"
fi
if ! command -v python3 &>/dev/null; then
    log_warning "python3 not found — falling back to plain go test output"
fi
echo ""

log_info "=== Format (gofmt) ==="
run_cmd "gofmt (no changes needed)" "
    cd '$REPO_ROOT'
    test -z \"\$(gofmt -l . 2>/dev/null)\"
"
echo ""

log_info "=== Static analysis & build ==="
run_cmd "go vet" "cd '$REPO_ROOT' && go vet ./..."
run_cmd "go build" "cd '$REPO_ROOT' && go build -o /dev/null ."
echo ""

log_info "=== Tests (all packages) ==="
GO_TEST_FAILED=false
if command -v python3 &>/dev/null; then
    (
        cd "$REPO_ROOT"
        go test -p 1 -parallel 1 -count=1 -timeout=5m -json ./... 2>&1
    ) | python3 -c "
import sys, json
failed = False
for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    try:
        d = json.loads(line)
        action  = d.get('Action', '')
        pkg     = d.get('Package', '')
        elapsed = d.get('Elapsed', 0)
        ms      = int(elapsed * 1000)
        if 'Test' not in d:
            if action == 'pass':
                print(f'  \033[0;32m✓\033[0m {pkg} ({ms}ms)')
            elif action == 'fail':
                print(f'  \033[0;31m✗\033[0m {pkg} ({ms}ms)')
                failed = True
    except json.JSONDecodeError:
        print(line)
sys.exit(1 if failed else 0)
" || { log_error "go test ./..."; FAILED_TESTS+=("go test ./..."); GO_TEST_FAILED=true; }
else
    if ! (cd "$REPO_ROOT" && go test -p 1 -parallel 1 -count=1 -timeout=5m -v ./...); then
        log_error "go test ./..."
        FAILED_TESTS+=("go test ./...")
        GO_TEST_FAILED=true
    fi
fi
if [ "$GO_TEST_FAILED" = false ]; then
    log_success "go test ./..."
fi
echo ""

echo "=========================================="
echo "  Summary"
echo "=========================================="
echo -e "${GREEN}Passed steps: $TESTS_PASSED${NC}"
echo -e "${RED}Failed steps: $TESTS_FAILED${NC}"
echo ""

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo -e "${RED}Failed:${NC}"
    for t in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗${NC} $t"
    done
    echo ""
    exit 1
fi

echo -e "${GREEN}All checks passed.${NC}"
echo ""
exit 0
