package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func prepareDataSourceCloudAccountCloudStack() map[string]interface{} {
	return map[string]interface{}{
		"name":    "test-cloudstack-account-1",
		"context": "project",
	}
}

func TestDataSourceCloudAccountCloudStackRead(t *testing.T) {
	d := dataSourceCloudAccountCloudStack().TestResourceData()
	_ = d.Set("name", "test-cloudstack-account-1")
	_ = d.Set("context", "project")

	ctx := context.Background()
	diags := dataSourceCloudAccountCloudStackRead(ctx, d, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cloudstack-account-id-1", d.Id())
	assert.Equal(t, "test-cloudstack-account-1", d.Get("name"))
}

func TestDataSourceCloudAccountCloudStackReadByID(t *testing.T) {
	d := dataSourceCloudAccountCloudStack().TestResourceData()
	_ = d.Set("id", "test-cloudstack-account-id-1")
	_ = d.Set("context", "project")

	ctx := context.Background()
	diags := dataSourceCloudAccountCloudStackRead(ctx, d, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cloudstack-account-id-1", d.Id())
}
