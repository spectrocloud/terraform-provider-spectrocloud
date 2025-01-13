package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseDataSourceAWSAccountSchema() *schema.ResourceData {
	d := dataSourceCloudAccountAws().TestResourceData()
	return d
}
func TestReadAWSAccountFuncName(t *testing.T) {
	d := prepareBaseDataSourceAWSAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-aws-account-1")
	_ = d.Set("context", "project")
	diags = dataSourceCloudAccountAwsRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadAWSAccountFuncID(t *testing.T) {
	d := prepareBaseDataSourceAWSAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("id", "test-aws-account-id-1")
	diags = dataSourceCloudAccountAwsRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadAWSAccountFuncNegative(t *testing.T) {
	d := prepareBaseDataSourceAWSAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-aws-account-1")
	diags = dataSourceCloudAccountAwsRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find aws cloud account")
}

func prepareBaseDataSourceAzureAccountSchema() *schema.ResourceData {
	d := dataSourceCloudAccountAzure().TestResourceData()
	return d
}
func TestReadAzureAccountFuncName(t *testing.T) {
	d := prepareBaseDataSourceAzureAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-azure-account-1")
	diags = dataSourceCloudAccountAzureRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadAzureAccountFuncID(t *testing.T) {
	d := prepareBaseDataSourceAzureAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("id", "test-azure-account-id-1")
	diags = dataSourceCloudAccountAzureRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadAzureAccountFuncNegative(t *testing.T) {
	d := prepareBaseDataSourceAzureAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-azure-account-1")
	diags = dataSourceCloudAccountAzureRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find azure cloud account")
}

func prepareBaseDataSourceGcpAccountSchema() *schema.ResourceData {
	d := dataSourceCloudAccountGcp().TestResourceData()
	return d
}
func TestReadGcpAccountFuncName(t *testing.T) {
	d := prepareBaseDataSourceGcpAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-gcp-account-1")
	diags = dataSourceCloudAccountGcpRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadGcpAccountFuncID(t *testing.T) {
	d := prepareBaseDataSourceGcpAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("id", "test-gcp-account-id-1")
	diags = dataSourceCloudAccountGcpRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadGcpAccountFuncNegative(t *testing.T) {
	d := prepareBaseDataSourceGcpAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-gcp-account-1")
	diags = dataSourceCloudAccountGcpRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find gcp cloud account")
}

func prepareBaseDataSourceVsphereAccountSchema() *schema.ResourceData {
	d := dataSourceCloudAccountVsphere().TestResourceData()
	return d
}
func TestReadVsphereAccountFuncName(t *testing.T) {
	d := prepareBaseDataSourceVsphereAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-vsphere-account-1")
	diags = dataSourceCloudAccountVsphereRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadVsphereAccountFuncID(t *testing.T) {
	d := prepareBaseDataSourceVsphereAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("id", "test-vsphere-account-id-1")
	diags = dataSourceCloudAccountVsphereRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadVsphereAccountFuncNegative(t *testing.T) {
	d := prepareBaseDataSourceVsphereAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-vsphere-account-1")
	diags = dataSourceCloudAccountVsphereRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find vsphere cloud account")
}

func prepareBaseDataSourceOpenstackAccountSchema() *schema.ResourceData {
	d := dataSourceCloudAccountOpenStack().TestResourceData()
	return d
}
func TestReadOpenstackAccountFuncName(t *testing.T) {
	d := prepareBaseDataSourceOpenstackAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-openstack-account-1")
	diags = dataSourceCloudAccountOpenStackRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadOpenstackAccountFuncID(t *testing.T) {
	d := prepareBaseDataSourceOpenstackAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("id", "test-openstack-account-id-1")
	diags = dataSourceCloudAccountOpenStackRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadOpenstackAccountFuncNegative(t *testing.T) {
	d := prepareBaseDataSourceOpenstackAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-openstack-account-1")
	diags = dataSourceCloudAccountOpenStackRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find openstack cloud account")
}

func prepareBaseDataSourceMaasAccountSchema() *schema.ResourceData {
	d := dataSourceCloudAccountMaas().TestResourceData()
	return d
}
func TestReadMaasAccountFuncName(t *testing.T) {
	d := prepareBaseDataSourceMaasAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-maas-account-1")
	diags = dataSourceCloudAccountMaasRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadMaasAccountFuncID(t *testing.T) {
	d := prepareBaseDataSourceMaasAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("id", "test-maas-account-id-1")
	diags = dataSourceCloudAccountMaasRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadMaasAccountFuncNegative(t *testing.T) {
	d := prepareBaseDataSourceMaasAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-maas-account-1")
	diags = dataSourceCloudAccountMaasRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find maas cloud account")
}

func prepareBaseDataSourceCustomAccountSchema() *schema.ResourceData {
	d := dataSourceCloudAccountCustom().TestResourceData()
	return d
}
func TestReadCustomAccountFuncName(t *testing.T) {
	d := prepareBaseDataSourceCustomAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-custom-account-1")
	_ = d.Set("cloud", "nutanix")
	diags = dataSourceCloudAccountCustomRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadCustomAccountFuncID(t *testing.T) {
	d := prepareBaseDataSourceCustomAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("id", "test-custom-account-id-1")
	_ = d.Set("cloud", "nutanix")
	diags = dataSourceCloudAccountCustomRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadCustomAccountFuncNegative(t *testing.T) {
	d := prepareBaseDataSourceCustomAccountSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-custom-account-1")
	_ = d.Set("cloud", "nutanix")
	diags = dataSourceCloudAccountCustomRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find cloud account")
}
