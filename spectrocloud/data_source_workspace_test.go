package spectrocloud

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceWorkspaceRead(t *testing.T) {

	resourceData := dataSourceWorkspace().TestResourceData()
	_ = resourceData.Set("name", "test-workspace")
	diags := dataSourceWorkspaceRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-workspace", resourceData.Get("name").(string))
	assert.NotEmpty(t, resourceData.Id())
}

func TestDataSourceWorkspaceRead_MissingName(t *testing.T) {
	resourceData := dataSourceWorkspace().TestResourceData()

	diags := dataSourceWorkspaceRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Empty(t, diags)
}
