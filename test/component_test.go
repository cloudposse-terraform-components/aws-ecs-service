package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/random"
)

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestBasic() {
	const component = "aurora-postgres/basic"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	clusterName := strings.ToLower(random.UniqueId())

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
	inputs := map[string]interface{}{
		"name":                "db",
		"database_name":       "postgres",
		"admin_user":          "postgres",
		"database_port":       5432,
		"publicly_accessible": true,
		"allowed_cidr_blocks": []string{"0.0.0.0/0"},
		"cluster_name": clusterName,
	}
	componentInstance, _ := s.DeployAtmosComponent(s.T(), component, stack, &inputs)
	assert.NotNil(s.T(), componentInstance)

	databaseName := atmos.Output(s.T(), componentInstance, "database_name")
	assert.Equal(s.T(), "postgres", databaseName)

	adminUsername := atmos.Output(s.T(), componentInstance, "admin_username")
	assert.Equal(s.T(), "postgres", adminUsername)

	delegatedDnsOptions := s.GetAtmosOptions("dns-delegated", stack, nil)
	delegatedDomainName := atmos.Output(s.T(), delegatedDnsOptions, "default_domain_name")
	delegatedDomainNZoneId := atmos.Output(s.T(), delegatedDnsOptions, "default_dns_zone_id")

	masterHostname := atmos.Output(s.T(), componentInstance, "master_hostname")
	expectedMasterHostname := fmt.Sprintf("%s-%s-writer.%s", inputs["name"], componentInstance.Vars["cluster_name"], delegatedDomainName)
	assert.Equal(s.T(), expectedMasterHostname, masterHostname)

	replicasHostname := atmos.Output(s.T(), componentInstance, "replicas_hostname")
	expectedReplicasHostname := fmt.Sprintf("%s-%s-reader.%s", inputs["name"], componentInstance.Vars["cluster_name"], delegatedDomainName)
	assert.Equal(s.T(), expectedReplicasHostname, replicasHostname)

	ssmKeyPaths := atmos.OutputList(s.T(), componentInstance, "ssm_key_paths")
	assert.Equal(s.T(), 7, len(ssmKeyPaths))

	kmsKeyArn := atmos.Output(s.T(), componentInstance, "kms_key_arn")
	assert.NotEmpty(s.T(), kmsKeyArn)

	allowedSecurityGroups := atmos.OutputList(s.T(), componentInstance, "allowed_security_groups")
	assert.Equal(s.T(), 0, len(allowedSecurityGroups))

	clusterIdentifier := atmos.Output(s.T(), componentInstance, "cluster_identifier")

	configMap := map[string]interface{}{}
	atmos.OutputStruct(s.T(), componentInstance, "config_map", &configMap)

	assert.Equal(s.T(), clusterIdentifier, configMap["cluster"])
	assert.Equal(s.T(), databaseName, configMap["database"])
	assert.Equal(s.T(), masterHostname, configMap["hostname"])
	assert.EqualValues(s.T(), inputs["database_port"], configMap["port"])
	assert.Equal(s.T(), adminUsername, configMap["username"])

	masterHostnameDNSRecord := aws.GetRoute53Record(s.T(), delegatedDomainNZoneId, masterHostname, "CNAME", awsRegion)
	assert.Equal(s.T(), *masterHostnameDNSRecord.ResourceRecords[0].Value, configMap["endpoint"])

	passwordSSMKey, ok := configMap["password_ssm_key"].(string)
	assert.True(s.T(), ok, "password_ssm_key should be a string")

	adminUserPassword := aws.GetParameter(s.T(), awsRegion, passwordSSMKey)

	dbUrl, ok := configMap["endpoint"].(string)
	assert.True(s.T(), ok, "endpoint should be a string")

	dbPort, ok := inputs["database_port"].(int)
	assert.True(s.T(), ok, "database_port should be an int")

	schemaExistsInRdsInstance := aws.GetWhetherSchemaExistsInRdsPostgresInstance(s.T(), dbUrl, int32(dbPort), adminUsername, adminUserPassword, databaseName)
	assert.True(s.T(), schemaExistsInRdsInstance)

	schemaExistsInRdsInstance = aws.GetWhetherSchemaExistsInRdsPostgresInstance(s.T(), masterHostname, int32(dbPort), adminUsername, adminUserPassword, databaseName)
	assert.True(s.T(), schemaExistsInRdsInstance)

	schemaExistsInRdsInstance = aws.GetWhetherSchemaExistsInRdsPostgresInstance(s.T(), replicasHostname, int32(dbPort), adminUsername, adminUserPassword, databaseName)
	assert.True(s.T(), schemaExistsInRdsInstance)

	s.DriftTest(component, stack, &inputs)
}

func (s *ComponentSuite) TestServerless() {
	const component = "aurora-postgres/serverless"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	clusterName := strings.ToLower(random.UniqueId())

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
	inputs := map[string]interface{}{
		"name":                "db",
		"database_name":       "postgres",
		"admin_user":          "postgres",
		"database_port":       5432,
		"publicly_accessible": true,
		"allowed_cidr_blocks": []string{"0.0.0.0/0"},
		"cluster_name": clusterName,
	}
	componentInstance, _ := s.DeployAtmosComponent(s.T(), component, stack, &inputs)
	assert.NotNil(s.T(), componentInstance)

	databaseName := atmos.Output(s.T(), componentInstance, "database_name")
	assert.Equal(s.T(), "postgres", databaseName)

	adminUsername := atmos.Output(s.T(), componentInstance, "admin_username")
	assert.Equal(s.T(), "postgres", adminUsername)

	delegatedDnsOptions := s.GetAtmosOptions("dns-delegated", stack, nil)
	delegatedDomainName := atmos.Output(s.T(), delegatedDnsOptions, "default_domain_name")
	delegatedDomainNZoneId := atmos.Output(s.T(), delegatedDnsOptions, "default_dns_zone_id")

	masterHostname := atmos.Output(s.T(), componentInstance, "master_hostname")
	expectedMasterHostname := fmt.Sprintf("%s-%s-writer.%s", inputs["name"], componentInstance.Vars["cluster_name"], delegatedDomainName)
	assert.Equal(s.T(), expectedMasterHostname, masterHostname)

	ssmKeyPaths := atmos.OutputList(s.T(), componentInstance, "ssm_key_paths")
	assert.Equal(s.T(), 7, len(ssmKeyPaths))

	kmsKeyArn := atmos.Output(s.T(), componentInstance, "kms_key_arn")
	assert.NotEmpty(s.T(), kmsKeyArn)

	allowedSecurityGroups := atmos.OutputList(s.T(), componentInstance, "allowed_security_groups")
	assert.Equal(s.T(), 0, len(allowedSecurityGroups))

	clusterIdentifier := atmos.Output(s.T(), componentInstance, "cluster_identifier")

	configMap := map[string]interface{}{}
	atmos.OutputStruct(s.T(), componentInstance, "config_map", &configMap)

	assert.Equal(s.T(), clusterIdentifier, configMap["cluster"])
	assert.Equal(s.T(), databaseName, configMap["database"])
	assert.Equal(s.T(), masterHostname, configMap["hostname"])
	assert.EqualValues(s.T(), inputs["database_port"], configMap["port"])
	assert.Equal(s.T(), adminUsername, configMap["username"])

	masterHostnameDNSRecord := aws.GetRoute53Record(s.T(), delegatedDomainNZoneId, masterHostname, "CNAME", awsRegion)
	assert.Equal(s.T(), *masterHostnameDNSRecord.ResourceRecords[0].Value, configMap["endpoint"])

	passwordSSMKey, ok := configMap["password_ssm_key"].(string)
	assert.True(s.T(), ok, "password_ssm_key should be a string")

	adminUserPassword := aws.GetParameter(s.T(), awsRegion, passwordSSMKey)

	dbUrl, ok := configMap["endpoint"].(string)
	assert.True(s.T(), ok, "endpoint should be a string")

	dbPort, ok := inputs["database_port"].(int)
	assert.True(s.T(), ok, "database_port should be an int")

	schemaExistsInRdsInstance := aws.GetWhetherSchemaExistsInRdsPostgresInstance(s.T(), dbUrl, int32(dbPort), adminUsername, adminUserPassword, databaseName)
	assert.True(s.T(), schemaExistsInRdsInstance)

	schemaExistsInRdsInstance = aws.GetWhetherSchemaExistsInRdsPostgresInstance(s.T(), masterHostname, int32(dbPort), adminUsername, adminUserPassword, databaseName)
	assert.True(s.T(), schemaExistsInRdsInstance)

	s.DriftTest(component, stack, &inputs)
}

func (s *ComponentSuite) TestDisabled() {
	const component = "aurora-postgres/disabled"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	s.VerifyEnabledFlag(component, stack, nil)
}

func TestRunSuite(t *testing.T) {
	suite := new(ComponentSuite)

	suite.AddDependency(t, "vpc", "default-test", nil)

	subdomain := strings.ToLower(random.UniqueId())
	inputs := map[string]interface{}{
		"zone_config": []map[string]interface{}{
			{
				"subdomain": subdomain,
				"zone_name": "components.cptest.test-automation.app",
			},
		},
	}
	suite.AddDependency(t, "dns-delegated", "default-test", &inputs)
	helper.Run(t, suite)
}
