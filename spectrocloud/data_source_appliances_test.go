package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseDataSourceAppliancesSchema() *schema.ResourceData {
	d := dataSourceAppliances().TestResourceData()
	d.SetId("test123")
	err := d.Set("context", "project")
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

func TestDataSourceAppliancesReadFunc(t *testing.T) {
	d := prepareBaseDataSourceAppliancesSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	diags = dataSourcesApplianceRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestDataSourceAppliancesReadNegativeFunc(t *testing.T) {
	d := prepareBaseDataSourceAppliancesSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	diags = dataSourcesApplianceRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "No edge host found")
}
