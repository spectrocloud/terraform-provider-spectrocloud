package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareDataSourceCloudAccountAzureSchema() *schema.ResourceData {
	d := dataSourceCloudAccountAzure().TestResourceData()
	return d
}

func TestDataSourceCloudAccountAzureRead(t *testing.T) {
	d := prepareDataSourceCloudAccountAzureSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-azure-account-1")
	diags = dataSourceCloudAccountAzureRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
	assert.Equal(t, "test", d.Get("tenant_name"))
	assert.Equal(t, false, d.Get("disable_properties_request"))
}

func TestDataSourceCloudAccountAzureReadNegative(t *testing.T) {
	d := prepareDataSourceCloudAccountAzureSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("name", "test-azure-account-1")
	diags = dataSourceCloudAccountAzureRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Unable to find azure cloud account")
}
