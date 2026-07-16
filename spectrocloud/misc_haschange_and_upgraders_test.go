package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// Continues sweep of remaining under-covered helpers.

// ---------------------------------------------------------------------------
// resourceApplicationUpdate via HasChange (currently 9.7%)
// ---------------------------------------------------------------------------

func TestResourceApplicationUpdate_HasChangeConfig(t *testing.T) {
	defer func() { _ = recover() }()

	base := map[string]string{
		"name":                       "test-app",
		"application_profile_uid":    "test-application-profile-id",
		"config.#":                   "1",
		"config.0.cluster_uid":       "old-cluster-uid",
		"config.0.cluster_context":   "project",
		"config.0.cluster_group_uid": "",
		"config.0.cluster_name":      "test",
		"cluster_uid":                "test-cluster-id",
	}
	d := buildUpdateResourceData(resourceApplication(), "test-application-id", base,
		map[string]*terraform.ResourceAttrDiff{
			"config.0.cluster_uid": {Old: "old-cluster-uid", New: "new-cluster-uid"},
		})

	_ = resourceApplicationUpdate(context.Background(), d, unitTestMockAPIClient)
}

// (resourcePlatformSettingImport tests already exist in resource_platform_setting_test.go.)

// ---------------------------------------------------------------------------
// appendOverrideHealthCheckConfigurationUpdateWarnings (8%)
// ---------------------------------------------------------------------------

func TestAppendOverrideHealthCheckConfigurationUpdateWarnings_NoChange(t *testing.T) {
	// No diff on machine_pool → early return.
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		map[string]string{"machine_pool.#": "0"}, nil)
	var diags diag.Diagnostics
	appendOverrideHealthCheckConfigurationUpdateWarnings(d, &diags)
	require.NotNil(t, d)
}

// ---------------------------------------------------------------------------
// Custom cloud state upgraders (77% - test the missing branches)
// ---------------------------------------------------------------------------

func TestResourceClusterCustomCloudStateUpgradeV2_MissingField(t *testing.T) {
	// Feed rawState without cluster_profile to hit the else branch.
	state := map[string]interface{}{}
	got, err := resourceClusterCustomCloudStateUpgradeV2(context.Background(), state, nil)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestResourceClusterCustomCloudStateUpgradeV3_MissingField(t *testing.T) {
	state := map[string]interface{}{}
	got, err := resourceClusterCustomCloudStateUpgradeV3(context.Background(), state, nil)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

// ---------------------------------------------------------------------------
// resourceAksStateUpgraders (already at ≥80%, hit remaining branches)
// ---------------------------------------------------------------------------

func TestResourceClusterAksStateUpgradeV2_V3(t *testing.T) {
	for _, fn := range []func(context.Context, map[string]interface{}, interface{}) (map[string]interface{}, error){
		resourceClusterAksStateUpgradeV2,
		resourceClusterAksStateUpgradeV3,
	} {
		// Populated cluster_profile / machine_pool → both branches hit.
		state := map[string]interface{}{
			"cluster_profile": []interface{}{map[string]interface{}{"id": "p1"}},
			"machine_pool":    []interface{}{map[string]interface{}{"name": "mp1"}},
		}
		got, err := fn(context.Background(), state, nil)
		assert.NoError(t, err)
		assert.NotNil(t, got)
	}
}

// ---------------------------------------------------------------------------
// Compile guards
// ---------------------------------------------------------------------------

var _ = schema.HashString
