package spectrocloud

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceRegistryPackRead(t *testing.T) {
	resourceData := dataSourceRegistryPack().TestResourceData()
	_ = resourceData.Set("name", "test-registry-name")
	diags := dataSourceRegistryPackRead(context.Background(), resourceData, unitTestMockAPIClient)
	assert.Equal(t, "test-registry-name", resourceData.Get("name").(string))
	assert.Empty(t, diags)
}

func TestDataSourceHelmRegistryPackRead(t *testing.T) {
	resourceData := dataSourceRegistryHelm().TestResourceData()
	_ = resourceData.Set("name", "Public")
	diags := dataSourceRegistryHelmRead(context.Background(), resourceData, unitTestMockAPIClient)
	assert.Equal(t, "Public", resourceData.Get("name").(string))
	assert.Empty(t, diags)
}

func TestDataSourceOciRegistryPackRead(t *testing.T) {
	resourceData := dataSourceRegistryOci().TestResourceData()
	_ = resourceData.Set("name", "test-registry-oci")
	diags := dataSourceRegistryOciRead(context.Background(), resourceData, unitTestMockAPIClient)
	assert.Equal(t, "test-registry-oci"+
		"", resourceData.Get("name").(string))
	assert.Empty(t, diags)
}

func TestDataSourceBasicRegistryPackRead(t *testing.T) {
	resourceData := dataSourceRegistry().TestResourceData()
	_ = resourceData.Set("name", "test-registry-name")
	diags := dataSourceRegistryRead(context.Background(), resourceData, unitTestMockAPIClient)
	assert.Equal(t, "test-registry-name", resourceData.Get("name").(string))
	assert.Empty(t, diags)
}
