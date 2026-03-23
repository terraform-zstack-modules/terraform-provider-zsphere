#!/bin/bash

# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TEST_DIR="${PROJECT_ROOT}/test/instance"

log_info() { echo -e "\033[0;32m[INFO]\033[0m $1"; }
log_warn() { echo -e "\033[0;33m[WARN]\033[0m $1"; }
log_error() { echo -e "\033[0;31m[ERROR]\033[0m $1"; }

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
        log_info "Please set the following environment variables:"
        echo "  export TF_VAR_zsphere_host=\"your-host\""
        echo "  export TF_VAR_zsphere_access_key_id=\"your-key-id\""
        echo "  export TF_VAR_zsphere_access_key_secret=\"your-secret\""
        echo ""
        return 1
    fi
    return 0
}

build_provider() {
    log_info "Building provider..."
    cd "${PROJECT_ROOT}"
    go build -o terraform-provider-zsphere

    local PLUGIN_DIR="${HOME}/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64"
    mkdir -p "${PLUGIN_DIR}"
    mv terraform-provider-zsphere "${PLUGIN_DIR}/"
    log_info "Provider installed to ${PLUGIN_DIR}"
}

terraform_init() {
    log_info "Running terraform init..."
    cd "${TEST_DIR}"
    terraform init -upgrade
}

run_tests() {
    cd "${TEST_DIR}"
    check_env_vars || return 1
    terraform_init
    log_info "Running all tests..."
    terraform test -verbose
}

run_basic_tests() {
    cd "${TEST_DIR}"
    check_env_vars || return 1
    terraform_init
    log_info "Running basic tests..."
    terraform test -filter=basic.tftest.hcl -verbose
}

run_crud_tests() {
    cd "${TEST_DIR}"
    check_env_vars || return 1
    terraform_init
    log_info "Running CRUD tests..."
    terraform test -filter=crud.tftest.hcl -verbose
}

run_validation_tests() {
    cd "${TEST_DIR}"
    check_env_vars || return 1
    terraform_init
    log_info "Running validation tests..."
    terraform test -filter=validation.tftest.hcl -verbose
}

run_specific_test() {
    cd "${TEST_DIR}"
    check_env_vars || return 1
    terraform_init
    log_info "Running specific test: $1"
    terraform test -run "$1" -verbose
}

clean_up() {
    log_info "Cleaning up test resources..."
    cd "${TEST_DIR}"
    rm -rf .terraform .terraform.lock.hcl terraform.tfstate* terraform.tfvars
    log_info "Cleanup completed"
}

show_help() {
    cat << EOF
Usage: $0 [command] [options]

Commands:
  build       Build and install the provider
  init        Run terraform init
  all         Run all tests
  basic       Run basic tests
  crud        Run CRUD tests
  validation  Run validation tests
  run <name>  Run a specific test by name
  clean       Clean up test resources
  help        Show this help message

Environment Variables:
  TF_VAR_zsphere_host              ZStack host address
  TF_VAR_zsphere_access_key_id     ZStack access key ID
  TF_VAR_zsphere_access_key_secret ZStack access key secret
  TF_VAR_image_uuid                Image UUID for testing
  TF_VAR_l3_network_uuid           L3 Network UUID for testing

Examples:
  $0 build
  $0 all
  $0 basic
  $0 run create_instance
EOF
}

main() {
    local command="${1:-help}"
    case "${command}" in
        build) build_provider ;;
        init) terraform_init ;;
        all) build_provider; run_tests ;;
        basic) build_provider; run_basic_tests ;;
        crud) build_provider; run_crud_tests ;;
        validation) build_provider; run_validation_tests ;;
        run) build_provider; run_specific_test "$2" ;;
        clean) clean_up ;;
        help|--help|-h) show_help ;;
        *) log_error "Unknown command: ${command}"; show_help; exit 1 ;;
    esac
}

main "$@"
