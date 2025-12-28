package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func prepareDataSourceCloudAccountApacheCloudStack() map[string]interface{} {
	return map[string]interface{}{
		"name":    "test-apache-cloudstack-account-1",
		"context": "project",
	}
}

func TestDataSourceCloudAccountApacheCloudStackRead(t *testing.T) {
	d := dataSourceCloudAccountApacheCloudStack().TestResourceData()
	_ = d.Set("name", "test-apache-cloudstack-account-1")
	_ = d.Set("context", "project")

	ctx := context.Background()
	diags := dataSourceCloudAccountApacheCloudStackRead(ctx, d, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-apache-cloudstack-account-id-1", d.Id())
	assert.Equal(t, "test-apache-cloudstack-account-1", d.Get("name"))
}

func TestDataSourceCloudAccountApacheCloudStackReadByID(t *testing.T) {
	d := dataSourceCloudAccountApacheCloudStack().TestResourceData()
	_ = d.Set("id", "test-apache-cloudstack-account-id-1")
	_ = d.Set("context", "project")

	ctx := context.Background()
	diags := dataSourceCloudAccountApacheCloudStackRead(ctx, d, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-apache-cloudstack-account-id-1", d.Id())
}
