## PR [ TODO ]

### Added

- Added `additional_lb_target_groups` variable to support registering multiple container ports to load balancer target groups
  - Enables sidecar containers (e.g., stunnel, envoy) to be registered to the same or separate target groups
  - `target_group_arn` is optional - defaults to using the same ALB/NLB target group as the service container
- Added `additional_security_groups` variable to attach additional security groups to the ECS service
- Added `security_group_id` field to `custom_security_group_rules` to allow rules to target different security groups

### Changed

- Updated `task_template` output to use S3 task definition when available via merge strategy
  - When `task-definition.json` exists in S3 (created by CI/CD), it takes precedence
  - Falls back to Terraform-created task definition when S3 version doesn't exist
  - Eliminates Terraform drift when CI/CD manages task definitions
- Fixed deprecation warnings:
  - Replaced `inline_policy` in `aws_iam_role.github_actions` with separate `aws_iam_role_policy` resource
  - Replaced deprecated `aws_s3_bucket_object` with `aws_s3_object`
- Updated SSM parameter key format to include attributes in the path for better organization

## PR [#1008](https://github.com/cloudposse/terraform-aws-components/pull/1008)

### Possible Breaking Change

- Refactored how S3 Task Definitions and the Terraform Task definition are merged.
  - Introduced local `local.containers_priority_terraform` to be referenced whenever terraform Should take priority
  - Introduced local `local.containers_priority_s3` to be referenced whenever S3 Should take priority
- `map_secrets` pulled out from container definition to local where it can be better maintained. Used Terraform as
  priority as it is a calculated as a map of arns.
- `s3_mirror_name` now automatically uploads a task-template.json to s3 mirror where it can be pulled from GitHub
  Actions.
