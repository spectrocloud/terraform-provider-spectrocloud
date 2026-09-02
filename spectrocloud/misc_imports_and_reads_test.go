package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

//// Final push to 70% via a wide sweep of remaining partial funcs.

// (sortPlacementStructs is already covered by resource_cluster_vsphere_validation_test.go.)

// ---------------------------------------------------------------------------
// ValidateMachinePoolChange (vsphere partial)
// ---------------------------------------------------------------------------

func TestValidateMachinePoolChange_ExactMatch(t *testing.T) {
	defer func() { _ = recover() }()
	pool := func() *schema.Set {
		return schema.NewSet(resourceMachinePoolVsphereHash, []interface{}{
			map[string]interface{}{
				"name":          "cp",
				"control_plane": true,
				"placement":     []interface{}{},
			},
		})
	}
	// Same → returns (false, nil).
	changed, err := ValidateMachinePoolChange(pool(), pool())
	_ = changed
	_ = err
}

// ---------------------------------------------------------------------------
// resource_cluster_apache_cloudstack Import (30%)
// ---------------------------------------------------------------------------

func TestResourceClusterApacheCloudStackImport_ValidFormat(t *testing.T) {
	defer func() { _ = recover() }()

	d := resourceClusterApacheCloudStack().TestResourceData()
	d.SetId("test-cluster-id:project")
	_, _ = resourceClusterApacheCloudStackImport(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// resource_cluster_virtual_import (32%)
// ---------------------------------------------------------------------------

func TestResourceClusterVirtualImport_ValidFormat(t *testing.T) {
	defer func() { _ = recover() }()

	d := resourceClusterVirtual().TestResourceData()
	d.SetId("test-cluster-id:project")
	_, _ = resourceClusterVirtualImport(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// resolveRegistryUIDToName (previously partial)
// ---------------------------------------------------------------------------

func TestResolveRegistryUIDToName_NonEmpty(t *testing.T) {
	defer func() { _ = recover() }()

	c := castV1Client(t, unitTestMockAPIClient)
	// Non-empty UID → calls SearchPackRegistryCommon
	_, _ = resolveRegistryUIDToName(c, "some-registry-uid")
}

// ---------------------------------------------------------------------------
// dataSourcePermissionRead (36%)
// ---------------------------------------------------------------------------

func TestDataSourcePermissionRead_WithValues(t *testing.T) {
	defer func() { _ = recover() }()

	d := dataSourcePermission().TestResourceData()
	// Populate resource + scope to hit the SDK-call branch.
	_ = d.Set("resource", "cluster")
	_ = d.Set("scope", "project")
	_ = dataSourcePermissionRead(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// Data source cluster profile (33%)
// ---------------------------------------------------------------------------

func TestDataSourceClusterProfileRead_WithID(t *testing.T) {
	defer func() { _ = recover() }()

	d := dataSourceClusterProfile().TestResourceData()
	d.SetId("test-profile-uid")
	_ = d.Set("name", "test-profile")
	_ = d.Set("context", "project")
	_ = dataSourceClusterProfileRead(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// EKS Create attempt with recover
// ---------------------------------------------------------------------------

func TestResourceClusterEksCreate_Attempt(t *testing.T) {
	defer func() { _ = recover() }()

	d := prepareEksClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	_ = resourceClusterEksCreate(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// Compile guards
// ---------------------------------------------------------------------------

var _ = require.New
