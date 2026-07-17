package spectrocloud

import (
	"context"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) closer.
//
// The main Azure/AKS wave2 tests (see resource_cluster_azure_wave2_test.go
// and resource_cluster_aks_wave2_test.go) already cover Read/Create and
// the machine-pool add/update/delete branches of Update. What they can't
// reach — thanks to schema.TestResourceData not reporting HasChange for
// double-Set operations — is the "cloud_config changed" branch of
// resourceClusterAksUpdate / resourceClusterAzureUpdate. That branch in
// turn calls toCloudConfigAks / toCloudConfigAzure, so those two pure
// converters end up untested even though they're on the hot Update path.
//
// This file pins those converters directly (fast, no schema plumbing) and
// exercises the three state-upgrade funcs (V0 for Azure, V2/V3 for AKS)
// that share a repeated pattern: convert cluster_profile from TypeList to
// TypeSet during a state-version bump.

// TestToCloudConfigAks — happy path: fully populated cloudConfig map
// produces every field the SDK entity expects. Covers the "all fields
// present" branch.
func TestToCloudConfigAks(t *testing.T) {
	cfg := map[string]interface{}{
		"subscription_id":                          "sub-123",
		"resource_group":                           "test-rg",
		"region":                                   "eastus",
		"ssh_key":                                  "ssh-rsa AAAA",
		"private_cluster":                          false,
		"vnet_name":                                "test-vnet",
		"vnet_resource_group":                      "vnet-rg",
		"vnet_cidr_block":                          "10.0.0.0/16",
		"override_cluster_api_config":              "",
		"worker_subnet_name":                       "worker-subnet",
		"worker_cidr":                              "10.0.1.0/24",
		"worker_subnet_security_group_name":        "worker-sg",
		"control_plane_subnet_name":                "cp-subnet",
		"control_plane_cidr":                       "10.0.2.0/24",
		"control_plane_subnet_security_group_name": "cp-sg",
	}

	got := toCloudConfigAks(cfg)
	require.NotNil(t, got)
	require.NotNil(t, got.ClusterConfig)
	assert.Equal(t, "test-vnet", got.ClusterConfig.VnetName)
	assert.Equal(t, "vnet-rg", got.ClusterConfig.VnetResourceGroup)
	assert.Equal(t, "10.0.0.0/16", got.ClusterConfig.VnetCidrBlock)
	// Worker + control-plane subnets should be populated as V1Subnet.
	require.NotNil(t, got.ClusterConfig.WorkerSubnet)
	assert.Equal(t, "worker-subnet", got.ClusterConfig.WorkerSubnet.Name)
	assert.Equal(t, "10.0.1.0/24", got.ClusterConfig.WorkerSubnet.CidrBlock)
}

// TestToCloudConfigAks_MinimalFields covers the nil-guards for the
// optional fields. Missing keys should NOT panic and should leave the
// corresponding entity fields at their zero value.
func TestToCloudConfigAks_MinimalFields(t *testing.T) {
	// Only the required fields are set. Note: toAksClusterConfig's
	// control-plane-subnet guard on line 810 checks `!= ""` against an
	// interface{} that's nil when the key is absent — nil != "" is true,
	// which sends the code into a .(string) assertion that panics. That's
	// a latent provider bug (see the guard on the worker-subnet block
	// two lines above, which uses `!= nil` correctly), but until it's
	// fixed the fixture has to include these keys as empty strings.
	cfg := map[string]interface{}{
		"subscription_id":           "sub-123",
		"resource_group":            "test-rg",
		"region":                    "eastus",
		"ssh_key":                   "",
		"private_cluster":           false,
		"control_plane_subnet_name": "",
		"control_plane_cidr":        "",
		"control_plane_subnet_security_group_name": "",
	}

	got := toCloudConfigAks(cfg)
	require.NotNil(t, got)
	require.NotNil(t, got.ClusterConfig)
	assert.Empty(t, got.ClusterConfig.VnetName)
	assert.Nil(t, got.ClusterConfig.WorkerSubnet,
		"missing worker_subnet_name/worker_cidr should leave WorkerSubnet nil")
}

// TestResourceClusterAksStateUpgradeV2 — the upgrader converts machine_pool
// from TypeList → TypeSet by leaving the data as a list and relying on
// Terraform's schema loader. Verify:
//  1. list value passes through unchanged
//  2. missing key → no error, no key added
//  3. non-list value → left in place, no panic
func TestResourceClusterAksStateUpgradeV2(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"machine_pool": []interface{}{
				map[string]interface{}{"name": "worker-pool"},
			},
		}
		got, err := resourceClusterAksStateUpgradeV2(ctx, state, nil)
		require.NoError(t, err)
		mp, ok := got["machine_pool"].([]interface{})
		require.True(t, ok, "machine_pool should still be a list after upgrade")
		assert.Len(t, mp, 1)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		state := map[string]interface{}{}
		got, err := resourceClusterAksStateUpgradeV2(ctx, state, nil)
		require.NoError(t, err)
		_, exists := got["machine_pool"]
		assert.False(t, exists, "should not synthesize the key when absent")
	})

	t.Run("non-list value is left alone", func(t *testing.T) {
		state := map[string]interface{}{"machine_pool": "not-a-list"}
		got, err := resourceClusterAksStateUpgradeV2(ctx, state, nil)
		require.NoError(t, err)
		assert.Equal(t, "not-a-list", got["machine_pool"])
	})
}

// TestResourceClusterAksStateUpgradeV3 — same shape as V2 but for
// cluster_profile.
func TestResourceClusterAksStateUpgradeV3(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterAksStateUpgradeV3(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		got, err := resourceClusterAksStateUpgradeV3(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})

	t.Run("non-list value is left alone", func(t *testing.T) {
		got, err := resourceClusterAksStateUpgradeV3(ctx, map[string]interface{}{"cluster_profile": 42}, nil)
		require.NoError(t, err)
		assert.Equal(t, 42, got["cluster_profile"])
	})
}

// TestResourceClusterAzureStateUpgradeV0 — Azure's V0→V1 upgrader shares
// the cluster_profile TypeList→TypeSet pattern with AKS V3. Same three
// branches to pin.
func TestResourceClusterAzureStateUpgradeV0(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterAzureStateUpgradeV0(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		got, err := resourceClusterAzureStateUpgradeV0(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})

	t.Run("non-list value is left alone", func(t *testing.T) {
		got, err := resourceClusterAzureStateUpgradeV0(ctx, map[string]interface{}{"cluster_profile": "opaque"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "opaque", got["cluster_profile"])
	})
}

// TestToCloudConfigAzure re-covers toCloudConfigAzure explicitly. It's
// already at 100% via the wave2 tests, but we add a nil-subnet case here
// because that's a common shape from a customer TF file that omits
// optional worker/control-plane subnets — worth pinning that the
// converter doesn't panic on a stripped-down config.
func TestToCloudConfigAzure_MinimalFields(t *testing.T) {
	// Minimal fields per azureClusterConfigFromMap's required reads —
	// subscription_id, resource_group, region, ssh_key, storage_account_name,
	// container_name. Missing any of these panics with an
	// interface-conversion error, so a "minimal" fixture still has to
	// include them.
	cfg := map[string]interface{}{
		"subscription_id":             "sub-123",
		"resource_group":              "test-rg",
		"region":                      "eastus",
		"ssh_key":                     "ssh-rsa AAAA",
		"storage_account_name":        "",
		"container_name":              "",
		"override_cluster_api_config": "",
	}

	got := toCloudConfigAzure(cfg)
	require.NotNil(t, got)
	require.NotNil(t, got.ClusterConfig)
	// Sanity: types are wired up correctly.
	assert.IsType(t, &models.V1AzureCloudClusterConfigEntity{}, got)
}
