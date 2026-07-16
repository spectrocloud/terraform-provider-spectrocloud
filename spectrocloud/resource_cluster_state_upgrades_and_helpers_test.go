package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) pure-function coverage for
// attachment / virtual / brownfield / (custom_cloud is already at 75%).
//
// Full CRUD tests aren't practical for these resources because each
// Create/Update path goes through waitForClusterCreation(..., initial=false)
// or waitForAddonDeploymentCreation, both of which bake in a 30-second
// initial Delay from Terraform's retry.StateChangeConf. That's fine for
// e2e tests but would balloon `make test` wall-clock for CI.
//
// Instead we hit the highest-ROI pure funcs and state upgraders:
//   - resourceAddonDeploymentStateUpgradeV2 (attachment)
//   - validateSingleClusterProfile (attachment)
//   - toBrownfieldClusterMetadata (brownfield)
//   - resourceClusterVirtualStateUpgradeV2 (virtual)
//   - resourceClusterBrownfieldStateUpgradeV1 (brownfield)
//   - flattenCloudConfigVirtual against the mock (virtual)

// ---------------------------------------------------------------------------
// attachment
// ---------------------------------------------------------------------------

func TestResourceAddonDeploymentStateUpgradeV2(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceAddonDeploymentStateUpgradeV2(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key no-op", func(t *testing.T) {
		got, err := resourceAddonDeploymentStateUpgradeV2(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})

	t.Run("non-list value left alone", func(t *testing.T) {
		got, err := resourceAddonDeploymentStateUpgradeV2(ctx, map[string]interface{}{"cluster_profile": "not-a-list"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "not-a-list", got["cluster_profile"])
	})
}

// TestValidateSingleClusterProfile pins the three branches of
// validateSingleClusterProfile: zero, one (ok), or many.
func TestValidateSingleClusterProfile(t *testing.T) {
	t.Run("zero profiles errors", func(t *testing.T) {
		d := resourceAddonDeployment().TestResourceData()
		err := validateSingleClusterProfile(d)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one cluster_profile is required")
	})

	t.Run("one profile ok", func(t *testing.T) {
		d := resourceAddonDeployment().TestResourceData()
		require.NoError(t, d.Set("cluster_profile", []interface{}{
			map[string]interface{}{"id": "cp-1"},
		}))
		assert.NoError(t, validateSingleClusterProfile(d))
	})

	t.Run("multiple profiles errors", func(t *testing.T) {
		d := resourceAddonDeployment().TestResourceData()
		require.NoError(t, d.Set("cluster_profile", []interface{}{
			map[string]interface{}{"id": "cp-1"},
			map[string]interface{}{"id": "cp-2"},
		}))
		err := validateSingleClusterProfile(d)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one cluster_profile is allowed")
	})
}

// TestGetAddonDeploymentIDHelpers pins the tiny ID pack/unpack helpers.
// They're 0.7% of the file but this is the code that decides which
// cluster + profile an addon deployment resource maps to during import.
// A break here silently corrupts import state.
func TestGetAddonDeploymentIDHelpers(t *testing.T) {
	assert.Equal(t, "cluster-abc", getClusterUID("cluster-abc_profile-def"))

	t.Run("valid two-part ID", func(t *testing.T) {
		got, err := getClusterProfileUID("cluster-abc_profile-def")
		require.NoError(t, err)
		assert.Equal(t, "profile-def", got)
	})

	t.Run("malformed one-part ID errors", func(t *testing.T) {
		_, err := getClusterProfileUID("no-separator")
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// brownfield
// ---------------------------------------------------------------------------

func TestResourceClusterBrownfieldStateUpgradeV1(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterBrownfieldStateUpgradeV1(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key no-op", func(t *testing.T) {
		got, err := resourceClusterBrownfieldStateUpgradeV1(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})
}

// TestToBrownfieldClusterMetadata is trivial — 4-line helper — but sits
// on the hot Create path for every brownfield import. A break would
// blank the cluster name silently.
func TestToBrownfieldClusterMetadata(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	require.NoError(t, d.Set("name", "my-imported-cluster"))

	got := toBrownfieldClusterMetadata(d)
	require.NotNil(t, got)
	assert.Equal(t, "my-imported-cluster", got.Name)
}

// ---------------------------------------------------------------------------
// virtual
// ---------------------------------------------------------------------------

func TestResourceClusterVirtualStateUpgradeV2(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterVirtualStateUpgradeV2(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key no-op", func(t *testing.T) {
		got, err := resourceClusterVirtualStateUpgradeV2(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})
}

// TestFlattenCloudConfigVirtualWithMock pins the cloud-config GET →
// flatten resources round-trip. Uses the VirtualClusterRoutes mock
// wired up in this batch.
func TestFlattenCloudConfigVirtualWithMock(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceClusterVirtual().TestResourceData()

	diags := flattenCloudConfigVirtual("test-cloud-config-id", d, c)
	assert.Empty(t, diags, "diags: %+v", diags)
	assert.Equal(t, "test-cloud-config-id", d.Get("cloud_config_id"))

	resources := d.Get("resources").([]interface{})
	require.Len(t, resources, 1)
	r := resources[0].(map[string]interface{})
	assert.Equal(t, 4, r["max_cpu"])
	assert.Equal(t, 8192, r["max_mem_in_mb"])
	assert.Equal(t, 40, r["max_storage_in_gb"])
	assert.Equal(t, 2, r["min_cpu"])
}

// TestFlattenCloudConfigVirtualNegative — when the config GET returns an
// error, the function should swallow it (log-only, empty diags) rather
// than propagating. This is the "graceful degradation" branch the code
// documents. Uses the negative client which has no virtual endpoint,
// so GetCloudConfigVirtual returns 404 → apiutil-not-found → nil, nil.
func TestFlattenCloudConfigVirtualNegative(t *testing.T) {
	c := mustUnitClient(t, true)
	d := resourceClusterVirtual().TestResourceData()

	diags := flattenCloudConfigVirtual("some-uid", d, c)
	assert.Empty(t, diags, "GET error must not propagate as diags")
	assert.Equal(t, "some-uid", d.Get("cloud_config_id"),
		"cloud_config_id still gets set even when the GET fails")
	// resources block should not be populated
	resources := d.Get("resources").([]interface{})
	assert.Empty(t, resources)
}

// Silence unused-var warning if a future edit removes schema import.
var _ = (*schema.ResourceData)(nil)
