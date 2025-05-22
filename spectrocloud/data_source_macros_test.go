package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func prepareMacrosTestData() *schema.ResourceData {
	d := dataSourceMacros().TestResourceData()
	d.Set("macros", map[string]interface{}{
		"macro1": "value1",
		"macro2": "value2",
	})
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
}

func TestDataSourceProjectMacrosRead(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareMacrosTestData()
	resourceData.Set("project", "Default")

	// Call the function
	diags := dataSourceMacrosRead(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
	assert.NotEmpty(t, resourceData.Id())
}
