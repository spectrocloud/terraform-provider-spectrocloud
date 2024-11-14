package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareDataSourceCloudAccountAwsSchema() *schema.ResourceData {
	d := dataSourceCloudAccountAws().TestResourceData()
	return d
}

func TestDataSourceCloudAccountAwsRead(t *testing.T) {
	d := prepareDataSourceCloudAccountAwsSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-aws-account-1")
	_ = d.Set("context", "project")
	diags = dataSourceCloudAccountAwsRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestDataSourceCloudAccountAwsReadNegative(t *testing.T) {
	d := prepareDataSourceCloudAccountAwsSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-aws-account-1")
	diags = dataSourceCloudAccountAwsRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find aws cloud account")
}
