#!/bin/bash

# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TEST_DIR="${PROJECT_ROOT}/test/image_storage"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_env_vars() {
    local missing_vars=()

    if [ -z "$TF_VAR_zsphere_host" ]; then
        missing_vars+=("TF_VAR_zsphere_host")
    fi
    if [ -z "$TF_VAR_zsphere_access_key_id" ]; then
        missing_vars+=("TF_VAR_zsphere_access_key_id")
    fi
    if [ -z "$TF_VAR_zsphere_access_key_secret" ]; then
        missing_vars+=("TF_VAR_zsphere_access_key_secret")
    fi

    if [ ${#missing_vars[@]} -gt 0 ]; then
        log_warn "Missing environment variables: ${missing_vars[*]}"
        log_info "Please set the following environment variables or create terraform.tfvars:"
        echo "  export TF_VAR_zsphere_host=\"your-host\""
        echo "  export TF_VAR_zsphere_access_key_id=\"your-key-id\""
        echo "  export TF_VAR_zsphere_access_key_secret=\"your-secret\""
        echo ""
        log_info "Or copy terraform.tfvars.example to terraform.tfvars and fill in the values"
    fi
}

build_provider() {
    log_info "Building terraform-provider-zsphere..."
    cd "${PROJECT_ROOT}"
    go build -o terraform-provider-zsphere

    log_info "Installing provider to local plugin directory..."
    local PLUGIN_DIR="${HOME}/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64"
    mkdir -p "${PLUGIN_DIR}"
    mv terraform-provider-zsphere "${PLUGIN_DIR}/"

    log_info "Provider installed successfully to ${PLUGIN_DIR}"
}

terraform_init() {
    log_info "Running terraform init..."
    cd "${TEST_DIR}"
    terraform init -upgrade
}

run_all_tests() {
    cd "${TEST_DIR}"
    check_env_vars
    terraform_init
    log_info "Running all tests..."
    terraform test -verbose
}

run_basic_tests() {
    cd "${TEST_DIR}"
    check_env_vars
    terraform_init
    log_info "Running basic tests..."
    terraform test -filter=basic.tftest.hcl -verbose
}

run_ceph_tests() {
    cd "${TEST_DIR}"
    check_env_vars
    terraform_init
    log_info "Running Ceph tests..."
    terraform test -filter=ceph.tftest.hcl -verbose
}

run_test_apply() {
    cd "${TEST_DIR}"
    check_env_vars
    terraform_init
    log_info "Running terraform apply..."
    terraform validate
    terraform apply -auto-approve
    log_info "Apply completed successfully"
}

run_test_destroy() {
    cd "${TEST_DIR}"
    log_info "Running terraform destroy..."
    terraform destroy -auto-approve
    log_info "Destroy completed successfully"
}

clean_up() {
    log_info "Cleaning up test resources..."
    cd "${TEST_DIR}"
    rm -rf .terraform .terraform.lock.hcl terraform.tfstate* terraform.tfvars
    log_info "Cleanup completed"
}

show_help() {
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  build       Build and install the provider"
    echo "  init        Run terraform init"
    echo "  all         Run all tests"
    echo "  basic       Run basic tests (ImageStore)"
    echo "  ceph        Run Ceph tests"
    echo "  apply       Run terraform apply"
    echo "  destroy     Run terraform destroy"
    echo "  clean       Clean up test resources"
    echo "  help        Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  TF_VAR_zsphere_host              ZStack host address"
    echo "  TF_VAR_zsphere_access_key_id     ZStack access key ID"
    echo "  TF_VAR_zsphere_access_key_secret ZStack access key secret"
    echo ""
    echo "Examples:"
    echo "  $0 build"
    echo "  $0 basic"
    echo "  $0 ceph"
    echo "  $0 apply"
}

main() {
    local command="${1:-help}"

    case "${command}" in
        build)
            build_provider
            ;;
        init)
            terraform_init
            ;;
        all)
            build_provider
            run_all_tests
            ;;
        basic)
            build_provider
            run_basic_tests
            ;;
        ceph)
            build_provider
            run_ceph_tests
            ;;
        apply)
            build_provider
            run_test_apply
            ;;
        destroy)
            run_test_destroy
            ;;
        clean)
            clean_up
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: ${command}"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
