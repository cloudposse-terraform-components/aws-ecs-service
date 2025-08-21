package test

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/config"
)

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestBasic() {
	const component = "ecs-service/echo-server"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	// Get the ECS cluster ARN from the dependency
	ecsOptions := s.GetAtmosOptions("ecs-cluster", stack, nil)
	clusterARN := atmos.Output(s.T(), ecsOptions, "ecs_cluster_arn")
	require.NotEmpty(s.T(), clusterARN, "ECS cluster ARN should not be empty")

	options, _ := s.DeployAtmosComponent(s.T(), component, stack, nil)
	assert.NotNil(s.T(), options)

	// Get the service name from the output
	serviceName := atmos.Output(s.T(), options, "service_name")
	require.NotEmpty(s.T(), serviceName, "Service name should not be empty")

	// Create AWS clients
	awsConfig, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
	require.NoError(s.T(), err, "Failed to load AWS config")

	ecsClient := ecs.NewFromConfig(awsConfig)
	cloudwatchClient := cloudwatchlogs.NewFromConfig(awsConfig)

	// Verify ECS service exists and is running
	service, err := ecsClient.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Cluster:  &clusterARN,
		Services: []string{serviceName},
	})
	assert.NoError(s.T(), err, "Failed to get ECS service %s", serviceName)
	assert.Len(s.T(), service.Services, 1, "Should find exactly one service")
	assert.Equal(s.T(), serviceName, *service.Services[0].ServiceName, "Service name should match")

	// Verify service is active
	assert.Equal(s.T(), "ACTIVE", *service.Services[0].Status, "Service should be active")

	// Get the task definition ARN from the output
	taskDefinitionARN := atmos.Output(s.T(), options, "task_definition_arn")
	require.NotEmpty(s.T(), taskDefinitionARN, "Task definition ARN should not be empty")

	// Verify task definition exists
	taskDef, err := ecsClient.DescribeTaskDefinition(context.Background(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefinitionARN,
	})
	assert.NoError(s.T(), err, "Failed to get task definition %s", taskDefinitionARN)
	assert.Equal(s.T(), taskDefinitionARN, *taskDef.TaskDefinition.TaskDefinitionArn, "Task definition ARN should match")

	// Verify CloudWatch log groups exist (if logs module is enabled)
	logsOutput := atmos.Output(s.T(), options, "logs")
	if logsOutput != "" {
		// Try to get log group names from the logs output
		// The exact structure depends on the cloudwatch-logs module
		logGroupName := atmos.Output(s.T(), options, "logs.log_group_name")
		if logGroupName != "" {
			_, err := cloudwatchClient.DescribeLogGroups(context.Background(), &cloudwatchlogs.DescribeLogGroupsInput{
				LogGroupNamePrefix: &logGroupName,
			})
			assert.NoError(s.T(), err, "Failed to get log group %s", logGroupName)
		}
	}

	// Verify ECS cluster exists and is accessible
	cluster, err := ecsClient.DescribeClusters(context.Background(), &ecs.DescribeClustersInput{
		Clusters: []string{clusterARN},
	})
	assert.NoError(s.T(), err, "Failed to get ECS cluster %s", clusterARN)
	assert.Len(s.T(), cluster.Clusters, 1, "Should find exactly one cluster")
	assert.Equal(s.T(), "ACTIVE", *cluster.Clusters[0].Status, "Cluster should be active")

	// Verify other important outputs
	ecsClusterARN := atmos.Output(s.T(), options, "ecs_cluster_arn")
	assert.NotEmpty(s.T(), ecsClusterARN, "ECS cluster ARN should not be empty")
	assert.Equal(s.T(), clusterARN, ecsClusterARN, "ECS cluster ARN should match dependency")

	vpcID := atmos.Output(s.T(), options, "vpc_id")
	assert.NotEmpty(s.T(), vpcID, "VPC ID should not be empty")

	subnetIDs := atmos.OutputList(s.T(), options, "subnet_ids")
	assert.NotEmpty(s.T(), subnetIDs, "Subnet IDs should not be empty")

	// Verify service image output
	serviceImage := atmos.Output(s.T(), options, "service_image")
	assert.NotEmpty(s.T(), serviceImage, "Service image should not be empty")

	s.DriftTest(component, stack, nil)
	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
}

func (s *ComponentSuite) TestEnabledFlag() {
	const component = "ecs-service/disabled"
	const stack = "default-test"
	s.VerifyEnabledFlag(component, stack, nil)
}

func (s *ComponentSuite) SetupSuite() {
	s.TestSuite.InitConfig()
	s.TestSuite.Config.ComponentDestDir = "components/terraform/ecs-service"
	s.TestSuite.SetupSuite()
}

func TestRunSuite(t *testing.T) {
	suite := new(ComponentSuite)

	suite.AddDependency(t, "vpc", "default-test", nil)

	randDomain := strings.ToLower(random.UniqueId())
	primaryDomainName := randDomain + ".components.cptest.test-automation.app"
	dnsPrimaryInputs := map[string]interface{}{
		"domain_names": []string{primaryDomainName},
	}
	suite.AddDependency(t, "dns-primary", "default-test", &dnsPrimaryInputs)

	subdomain := strings.ToLower(random.UniqueId())
	dnsDelegatedInputs := map[string]interface{}{
		"zone_config": []map[string]interface{}{
			{
				"subdomain": subdomain,
				"zone_name": primaryDomainName,
			},
		},
	}
	suite.AddDependency(t, "dns-delegated", "default-test", &dnsDelegatedInputs)

	suite.AddDependency(t, "ecs-cluster", "default-test", nil)
	helper.Run(t, suite)
}
