package spectrocloud

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceClusterGroupRead_SystemContext(t *testing.T) {

	resourceData := dataSourceClusterGroup().TestResourceData()
	_ = resourceData.Set("name", "test-cluster-group")
	_ = resourceData.Set("context", "system")
	diags := dataSourceClusterGroupRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cluster-group", resourceData.Get("name").(string))
	assert.NotEmpty(t, resourceData.Id())
}

func TestDataSourceClusterGroupRead_TenantContext(t *testing.T) {
	resourceData := dataSourceClusterGroup().TestResourceData()
	_ = resourceData.Set("name", "test-cluster-group")
	_ = resourceData.Set("context", "tenant")

	diags := dataSourceClusterGroupRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cluster-group", resourceData.Get("name").(string))
	assert.NotEmpty(t, resourceData.Id())
}

func TestDataSourceClusterGroupRead_ProjectContext(t *testing.T) {
	resourceData := dataSourceClusterGroup().TestResourceData()
	_ = resourceData.Set("name", "test-cluster-group")
	_ = resourceData.Set("context", "project")

	diags := dataSourceClusterGroupRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cluster-group", resourceData.Get("name").(string))
	assert.NotEmpty(t, resourceData.Id())
}

func TestDataSourceClusterGroupRead_InvalidContext(t *testing.T) {
	resourceData := dataSourceClusterGroup().TestResourceData()
	_ = resourceData.Set("name", "test-cluster-group")
	_ = resourceData.Set("context", "other")

	diags := dataSourceClusterGroupRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "", resourceData.Id())
}
