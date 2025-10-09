#!/bin/bash

#############################################
# Delivery Management System - Test Runner
# Comprehensive test suite execution script
#############################################

set -e  # Exit on error
set -o pipefail  # Catch errors in pipes

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="${PROJECT_ROOT}/tests"
COVERAGE_DIR="${PROJECT_ROOT}/coverage"
COVERAGE_FILE="${COVERAGE_DIR}/coverage.out"
COVERAGE_HTML="${COVERAGE_DIR}/coverage.html"

# Test categories
UNIT_TESTS="unit_test.go"
INTEGRATION_TESTS="integration_test.go"
SERVICE_TESTS="user_service_test.go order_service_test.go"
HANDLER_TESTS="user_test.go order_test.go"
MIDDLEWARE_TESTS="middleware_test.go"
EDGE_CASE_TESTS="edge_case_test.go"
LOAD_TESTS="load_test.go"
BENCHMARK_TESTS="benchmark_test.go"

# Functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    print_success "Go $(go version | awk '{print $3}')"
    
    if ! command -v docker &> /dev/null; then
        print_warning "Docker not found - integration tests may fail"
    else
        print_success "Docker $(docker --version | awk '{print $3}' | tr -d ',')"
    fi
    
    # Check if test dependencies are available
    cd "$PROJECT_ROOT"
    if ! go list -m github.com/stretchr/testify &> /dev/null; then
        print_info "Installing test dependencies..."
        go get github.com/stretchr/testify
    fi
    print_success "Test dependencies available"
}

# Setup test environment
setup_test_env() {
    print_header "Setting Up Test Environment"
    
    # Create coverage directory
    mkdir -p "$COVERAGE_DIR"
    print_success "Coverage directory created"
    
    # Clean previous coverage data
    rm -f "$COVERAGE_FILE" "$COVERAGE_HTML"
    print_success "Previous coverage data cleaned"
}

# Run specific test category
run_test_category() {
    local category=$1
    local files=$2
    local extra_flags=$3
    
    print_header "Running $category"
    
    cd "$TEST_DIR"
    
    if [ -n "$files" ]; then
        for file in $files; do
            if [ -f "$file" ]; then
                print_info "Testing: $file"
                if go test -v $extra_flags "./$file" 2>&1 | tee /tmp/test_output.log; then
                    print_success "$file passed"
                else
                    print_error "$file failed"
                    return 1
                fi
            else
                print_warning "$file not found, skipping"
            fi
        done
    else
        if go test -v $extra_flags ./... 2>&1 | tee /tmp/test_output.log; then
            print_success "$category passed"
        else
            print_error "$category failed"
            return 1
        fi
    fi
    
    return 0
}

# Run all tests with coverage
run_all_tests_with_coverage() {
    print_header "Running All Tests with Coverage"
    
    cd "$PROJECT_ROOT"
    
    if go test ./tests/... -v -short -count=1 -coverprofile="$COVERAGE_FILE" -covermode=atomic 2>&1 | tee /tmp/test_output.log; then
        print_success "All tests passed"
        
        # Generate coverage report
        if [ -f "$COVERAGE_FILE" ]; then
            print_info "Generating coverage report..."
            go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"
            
            # Calculate coverage percentage
            COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
            print_success "Coverage: $COVERAGE"
            print_info "HTML report: $COVERAGE_HTML"
        fi
        return 0
    else
        print_error "Some tests failed"
        return 1
    fi
}

# Run race detector
run_race_detector() {
    print_header "Running Race Condition Detection"
    
    cd "$PROJECT_ROOT"
    
    if go test ./tests/... -race -short -count=1 2>&1 | tee /tmp/race_output.log; then
        print_success "No race conditions detected"
        return 0
    else
        print_error "Race conditions detected"
        return 1
    fi
}

# Run benchmarks
run_benchmarks() {
    print_header "Running Benchmarks"
    
    cd "$TEST_DIR"
    
    if [ -f "$BENCHMARK_TESTS" ]; then
        if go test -bench=. -benchmem "./$BENCHMARK_TESTS" 2>&1 | tee /tmp/benchmark_output.log; then
            print_success "Benchmarks completed"
            return 0
        else
            print_warning "Benchmarks had issues"
            return 1
        fi
    else
        print_warning "Benchmark tests not found"
        return 0
    fi
}

# Generate test summary
generate_summary() {
    print_header "Test Summary"
    
    if [ -f "$COVERAGE_FILE" ]; then
        COVERAGE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
        if [ -n "$COVERAGE" ] && [ "$(echo "$COVERAGE > 0" | bc -l 2>/dev/null || echo 0)" -eq 1 ]; then
            echo -e "${BLUE}Coverage Report:${NC}"
            go tool cover -func="$COVERAGE_FILE" | tail -10
        fi
    fi
    
    if [ -f /tmp/test_output.log ]; then
        # Count only top-level tests (exclude subtests with /)
        TOTAL_TESTS=$(grep "^=== RUN" /tmp/test_output.log | grep -v "/" | wc -l | tr -d ' ')
        PASSED_TESTS=$(grep "^--- PASS" /tmp/test_output.log | grep -v "/" | wc -l | tr -d ' ')
        FAILED_TESTS=$(grep "^--- FAIL" /tmp/test_output.log | grep -v "/" | wc -l | tr -d ' ')
        SKIPPED_TESTS=$(grep "^--- SKIP" /tmp/test_output.log | grep -v "/" | wc -l | tr -d ' ')
        
        echo -e "\n${BLUE}Test Statistics:${NC}"
        echo -e "  Total:   $TOTAL_TESTS"
        echo -e "  ${GREEN}Passed:  $PASSED_TESTS${NC}"
        if [ "$FAILED_TESTS" -gt 0 ]; then
            echo -e "  ${RED}Failed:  $FAILED_TESTS${NC}"
        else
            echo -e "  Failed:  $FAILED_TESTS"
        fi
        if [ "$SKIPPED_TESTS" -gt 0 ]; then
            echo -e "  ${YELLOW}Skipped: $SKIPPED_TESTS${NC}"
        else
            echo -e "  Skipped: $SKIPPED_TESTS"
        fi
    fi
}

# Main execution
main() {
    local test_mode=${1:-"all"}
    local exit_code=0
    
    print_header "Delivery Management System - Test Suite"
    print_info "Mode: $test_mode"
    
    check_prerequisites
    setup_test_env
    
    case "$test_mode" in
        "unit")
            run_test_category "Unit Tests" "$UNIT_TESTS" "" || exit_code=1
            ;;
        "integration")
            run_test_category "Integration Tests" "$INTEGRATION_TESTS" "-timeout 30s" || exit_code=1
            ;;
        "service")
            run_test_category "Service Tests" "$SERVICE_TESTS" "" || exit_code=1
            ;;
        "handler")
            run_test_category "Handler Tests" "$HANDLER_TESTS" "" || exit_code=1
            ;;
        "middleware")
            run_test_category "Middleware Tests" "$MIDDLEWARE_TESTS" "" || exit_code=1
            ;;
        "edge")
            run_test_category "Edge Case Tests" "$EDGE_CASE_TESTS" "" || exit_code=1
            ;;
        "load")
            run_test_category "Load Tests" "$LOAD_TESTS" "-timeout 5m" || exit_code=1
            ;;
        "race")
            run_race_detector || exit_code=1
            ;;
        "bench")
            run_benchmarks || exit_code=1
            ;;
        "coverage")
            run_all_tests_with_coverage || exit_code=1
            ;;
        "all")
            run_all_tests_with_coverage || exit_code=1
            run_race_detector || exit_code=1
            run_benchmarks || exit_code=1
            ;;
        "quick")
            run_test_category "Quick Tests" "$UNIT_TESTS $EDGE_CASE_TESTS" "-short" || exit_code=1
            ;;
        *)
            echo "Usage: $0 {all|unit|integration|service|handler|middleware|edge|load|race|bench|coverage|quick}"
            echo ""
            echo "Test modes:"
            echo "  all         - Run all tests with coverage, race detection, and benchmarks"
            echo "  unit        - Run unit tests only"
            echo "  integration - Run integration tests only"
            echo "  service     - Run service layer tests"
            echo "  handler     - Run handler tests"
            echo "  middleware  - Run middleware tests"
            echo "  edge        - Run edge case tests"
            echo "  load        - Run load tests"
            echo "  race        - Run race condition detection"
            echo "  bench       - Run benchmarks"
            echo "  coverage    - Run all tests with coverage report"
            echo "  quick       - Run quick tests (unit + edge cases)"
            exit 1
            ;;
    esac
    
    generate_summary
    
    if [ $exit_code -eq 0 ]; then
        print_header "All Tests Completed Successfully! ✓"
    else
        print_header "Some Tests Failed! ✗"
    fi
    
    exit $exit_code
}

# Run main function
main "$@"
