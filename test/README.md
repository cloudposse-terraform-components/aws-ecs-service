# Tests

## Overview

- Framework: Go Terratest + Cloud Posse test-helpers.
- Orchestration: Atmos (stacks + vendored components).
- Scope: Deploys minimal dependencies (VPC, DNS, ECS cluster) and the target ECS service component, then performs assertions and drift checks.

## How It Works

- Test suite: `test/component_test.go` using `github.com/cloudposse/test-helpers/pkg/atmos/component-helper`.
- Stacks: `test/fixtures/stacks` define catalog entries for dependencies and use cases.
- Vendoring: `test/fixtures/vendor.yaml` lists upstream components to fetch locally for tests.
- Backend: Local state paths are configured in `test/fixtures/stacks/orgs/default/test/_defaults.yaml`.

## Run Via ChatOps

- Open a Pull Request with your changes.
- Comment on the PR with `/terratest`.
- A CI workflow will:
  - Vendor the fixtures (`atmos vendor pull -f test/fixtures/vendor.yaml`).
  - Run the Go tests in `test/` against the fixtures.
- Monitor the PR checks for the Terratest job result.

## Test Inputs and Env Vars

- Local backend state root can be overridden with `COMPONENT_HELPER_STATE_DIR`.
- Test account ID can be provided via `TEST_ACCOUNT_ID` and is used by `account-map` in `_defaults.yaml`.

## Notable Scenarios

- Basic service deploys nginx on Fargate without a load balancer, and applies component-level SG rules (`custom_security_group_rules`).
- Disabled scenario verifies `enabled=false` bypasses provisioning.

## File Map

- `test/component_test.go`: Test suite and cases.
- `test/fixtures/vendor.yaml`: Vendored components for dependencies.
- `test/fixtures/stacks/catalog`: Catalog entries for VPC, DNS, ECS cluster, and ECS service use cases.
- `test/fixtures/stacks/orgs/default/test/_defaults.yaml`: Shared test settings (local backend, account map, labels).
- `test/fixtures/stacks/orgs/default/test/tests.yaml`: Imports catalog entries to compose the test stack.

## Troubleshooting

- Ensure vendored components exist under `components/terraform/*` after running `atmos vendor pull`.
- If tests fail looking up remote state, verify dependencies (VPC, DNS, ECS cluster) are included and enabled in the stack.
- For flaky infra issues, re-run `/terratest` on the PR or `go test` locally.
