package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func prepareMacrosTestData() *schema.ResourceData {
	d := dataSourceMacros().TestResourceData()
	d.Set("context", "tenant")
	return d
}

func TestDataSourceMacrosRead(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareMacrosTestData()

	// Call the function
	diags := dataSourceMacrosRead(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
	assert.NotEmpty(t, resourceData.Id())

	// Verify macros are set
	macros, ok := resourceData.GetOk("macros_map")
	assert.True(t, ok)
	macroMap, ok := macros.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, macroMap)
}

func TestDataSourceProjectMacrosRead(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareMacrosTestData()
	resourceData.Set("context", "project")

	// Call the function
	diags := dataSourceMacrosRead(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
	assert.NotEmpty(t, resourceData.Id())

	// Verify macros are set
	macros, ok := resourceData.GetOk("macros_map")
	assert.True(t, ok)
	macroMap, ok := macros.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, macroMap)
}
