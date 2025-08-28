package test

import (
	"strings"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestBasic() {
	const component = "ecs-service/basic"
	const stack = "default-test"

	serviceName := strings.ToLower(random.UniqueId())

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)

	inputs := map[string]interface{}{
		"name": serviceName,
		"containers": map[string]interface{}{
			"service": map[string]interface{}{
				"name":  "nginx",
				"image": "nginx:stable",
				"port_mappings": []map[string]interface{}{
					{"containerPort": 80, "hostPort": 80, "protocol": "tcp"},
				},
			},
		},
		"task": map[string]interface{}{
			"launch_type":   "FARGATE",
			"network_mode":  "awsvpc",
			"desired_count": 1,
		},
		// No LB for simpler test surface
		"use_lb": false,

		// Simple CIDR-based ingress rule to exercise new feature
		"custom_security_group_rules": []map[string]interface{}{
			{
				"type":        "ingress",
				"protocol":    "tcp",
				"from_port":   80,
				"to_port":     80,
				"cidr_blocks": []string{"0.0.0.0/0"},
				"description": "Allow HTTP from anywhere",
			},
		},
	}

	componentInstance, _ := s.DeployAtmosComponent(s.T(), component, stack, &inputs)
	require.NotNil(s.T(), componentInstance)
	// Basic smoke outputs
	clusterArn := atmos.Output(s.T(), componentInstance, "ecs_cluster_arn")
	assert.NotEmpty(s.T(), clusterArn)

	subnets := atmos.OutputList(s.T(), componentInstance, "subnet_ids")
	assert.GreaterOrEqual(s.T(), len(subnets), 1)

	serviceNameOut := atmos.Output(s.T(), componentInstance, "service_name")
	assert.Contains(s.T(), serviceNameOut, serviceName)
	// Drift test ensures idempotency
	s.DriftTest(component, stack, &inputs)
}

func (s *ComponentSuite) TestDisabled() {
	const component = "ecs-service/disabled"
	const stack = "default-test"

	s.VerifyEnabledFlag(component, stack, nil)
}

func TestRunSuite(t *testing.T) {
	suite := new(ComponentSuite)

	// Dependencies: VPC, DNS delegated (for service domain), and ECS cluster
	suite.AddDependency(t, "vpc", "default-test", nil)

	// Minimal DNS delegated zone so service domain lookups resolve
	subdomain := strings.ToLower(random.UniqueId())
	dnsInputs := map[string]interface{}{
		"zone_config": []map[string]interface{}{
			{
				"subdomain": subdomain,
				"zone_name": "components.cptest.test-automation.app",
			},
		},
	}
	suite.AddDependency(t, "dns-delegated", "default-test", &dnsInputs)

	// ECS cluster dependency
	ecsInputs := map[string]interface{}{
		"name":                   "cluster",
		"acm_certificate_domain": subdomain + "components.cptest.test-automation.app",
	}
	suite.AddDependency(t, "ecs-cluster", "default-test", &ecsInputs)

	helper.Run(t, suite)
}
