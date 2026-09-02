package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDataSourceClusterConfigPolicyRead exercises the success path against
// the mock's getClusterConfigPolicyResponse() fixture
// (tests/mockApiServer/routes/mockClusterConfigPolicy.go), which returns a
// policy named "test-cluster-config-policy" with UID
// "test-cluster-config-policy-id".
func TestDataSourceClusterConfigPolicyRead(t *testing.T) {
	d := dataSourceClusterConfigPolicy().TestResourceData()
	_ = d.Set("name", "test-cluster-config-policy")
	_ = d.Set("context", "project")

	diags := dataSourceClusterConfigPolicyRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-cluster-config-policy-id", d.Id())
	assert.Equal(t, "test-cluster-config-policy", d.Get("name"))
	schedules := d.Get("schedules").([]interface{})
	assert.Len(t, schedules, 1)
}

// TestDataSourceClusterConfigPolicyRead_NotFound exercises the
// GetClusterConfigPolicyByName error branch: the mock's policy list only
// contains "test-cluster-config-policy", so any other name is filtered out
// client-side and the SDK returns a clean "not found" error.
func TestDataSourceClusterConfigPolicyRead_NotFound(t *testing.T) {
	d := dataSourceClusterConfigPolicy().TestResourceData()
	_ = d.Set("name", "does-not-exist")
	_ = d.Set("context", "project")

	diags := dataSourceClusterConfigPolicyRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
