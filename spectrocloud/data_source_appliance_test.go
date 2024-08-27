package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseDataSourceApplianceSchema() *schema.ResourceData {
	d := dataSourceAppliance().TestResourceData()
	d.SetId("test123")
	err := d.Set("name", "test-edge-01")
	if err != nil {
		return nil
	}
	err = d.Set("tags", map[string]string{"test": "true"})
	if err != nil {
		return nil
	}
	err = d.Set("status", "ready")
	if err != nil {
		return nil
	}
	err = d.Set("health", "healthy")
	if err != nil {
		return nil
	}
	err = d.Set("architecture", "amd")
	if err != nil {
		return nil
	}
	return d
}

func TestDataSourceApplianceReadFunc(t *testing.T) {
	d := prepareBaseDataSourceApplianceSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	diags = dataSourceApplianceRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestDataSourceApplianceReadNegativeFunc(t *testing.T) {
	d := prepareBaseDataSourceApplianceSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	diags = dataSourceApplianceRead(ctx, d, unitTestMockAPINegativeClient)
	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, "No edge host found", "The first diagnostic message does not contain the expected error message")
	}
}
