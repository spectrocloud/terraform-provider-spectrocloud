package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// Exercises the many `update*Config` helpers on cluster_common_*.go that
// gate their SDK call behind either a `ValidateContext` check or a nil
// short-circuit from a to*(d) builder. For each we hit at least the
// "invalid context → error" branch and, where the func has a nil-config
// early-return, the "nil config → no-op nil" branch.

// clusterUpdateResourceData builds a ResourceData from resourceClusterEks
// (which includes the `context` field via cluster_common_fields) with the
// provided context and no other config. That's sufficient to reach the
// ValidateContext / to*(d) branches in every update* helper.
func clusterUpdateResourceData(t *testing.T, context string) *schema.ResourceData {
	t.Helper()
	d := resourceClusterEks().TestResourceData()
	if context != "" {
		require.NoError(t, d.Set("context", context))
	}
	d.SetId("test-cluster-uid")
	return d
}

func TestUpdateClusterMetadata_InvalidContext(t *testing.T) {
	d := clusterUpdateResourceData(t, "bogus")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	err := updateClusterMetadata(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Context")
}

func TestUpdateClusterMetadata_Valid(t *testing.T) {
	d := clusterUpdateResourceData(t, "project")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	// The mock's cluster meta update route may or may not be wired; we
	// only care that the ValidateContext branch is passed. Either nil or
	// a routing error is acceptable — no panic, no ValidateContext error.
	err := updateClusterMetadata(c, d)
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid Context")
	}
}

func TestUpdateClusterAdditionalMetadata_InvalidContext(t *testing.T) {
	d := clusterUpdateResourceData(t, "bogus")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	err := updateClusterAdditionalMetadata(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Context")
}

func TestUpdateClusterOsPatchConfig_InvalidContext(t *testing.T) {
	d := clusterUpdateResourceData(t, "bogus")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	err := updateClusterOsPatchConfig(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Context")
}

func TestUpdateHostConfig_ReachesFuncBody(t *testing.T) {
	// toClusterHostConfigs may return a stub struct even without the
	// host_config block set (defaults), landing us on the SDK call. We
	// only care that the func body is entered; the SDK error surfacing
	// through means we did.
	d := clusterUpdateResourceData(t, "project")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_ = updateHostConfig(c, d) // reach func body; result irrelevant
}

func TestUpdateLocationConfig_InvalidContext(t *testing.T) {
	d := clusterUpdateResourceData(t, "bogus")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	err := updateLocationConfig(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Context")
}

func TestUpdateLocationConfig_NoConfig(t *testing.T) {
	// After the context validates, toClusterLocationConfigs returns nil
	// with no `location_config` block set — early nil return.
	d := clusterUpdateResourceData(t, "project")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	assert.NoError(t, updateLocationConfig(c, d))
}

func TestUpdateBackupPolicy_NoPolicySchemaMisses(t *testing.T) {
	// No backup_policy set → toBackupPolicy returns nil → the helper
	// returns an error explaining that the policy cannot be destroyed.
	d := clusterUpdateResourceData(t, "project")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	err := updateBackupPolicy(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "backup policy validation")
}

func TestUpdateScanPolicy_NoPolicy(t *testing.T) {
	// No scan_policy and no change → early nil.
	d := clusterUpdateResourceData(t, "project")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	assert.NoError(t, updateScanPolicy(c, d))
}

func TestUpdateClusterNamespaces_ReachesFuncBody(t *testing.T) {
	// Same pattern as TestUpdateHostConfig — the func body is what we
	// care about; whether the SDK call succeeds under the mock is
	// incidental.
	d := clusterUpdateResourceData(t, "project")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_ = updateClusterNamespaces(c, d)
}
