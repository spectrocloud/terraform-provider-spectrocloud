package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToDeveloperSetting(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	// Set custom values
	d.Set("virtual_clusters_limit", int32(10))
	d.Set("cpu", int32(4))
	d.Set("memory", int32(16))
	d.Set("storage", int32(50))

	devCredit, sysClusterGroupPref := toDeveloperSetting(d)

	assert.NotNil(t, devCredit)
	assert.NotNil(t, sysClusterGroupPref)
	assert.Equal(t, int32(10), devCredit.VirtualClustersLimit)
	assert.Equal(t, int32(4), devCredit.CPU)
	assert.Equal(t, int32(16), devCredit.MemoryGiB)
	assert.Equal(t, int32(50), devCredit.StorageGiB)
	assert.False(t, sysClusterGroupPref.HideSystemClusterGroups)
}

func TestToDeveloperSettingDefault(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	devCredit, sysClusterGroupPref := toDeveloperSettingDefault(d)

	assert.NotNil(t, devCredit)
	assert.NotNil(t, sysClusterGroupPref)
	assert.Equal(t, int32(12), devCredit.CPU)
	assert.Equal(t, int32(16), devCredit.MemoryGiB)
	assert.Equal(t, int32(20), devCredit.StorageGiB)
	assert.Equal(t, int32(2), devCredit.VirtualClustersLimit)
	assert.False(t, sysClusterGroupPref.HideSystemClusterGroups)
}

func TestFlattenDeveloperSetting(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	devSetting := &models.V1DeveloperCredit{
		CPU:                  8,
		MemoryGiB:            32,
		StorageGiB:           100,
		VirtualClustersLimit: 5,
	}
	sysClusterGroupPref := &models.V1TenantEnableClusterGroup{
		HideSystemClusterGroups: true,
	}

	err := flattenDeveloperSetting(devSetting, sysClusterGroupPref, d)
	assert.NoError(t, err)

	// Verify values set in schema
	assert.Equal(t, 8, d.Get("cpu"))
	assert.Equal(t, 32, d.Get("memory"))
	assert.Equal(t, 100, d.Get("storage"))
	assert.Equal(t, 5, d.Get("virtual_clusters_limit"))
	assert.True(t, d.Get("hide_system_cluster_group").(bool))
}

// ---------------------------------------------------------------------------
// CRUD coverage
// ---------------------------------------------------------------------------

func prepareDeveloperSettingResourceData() *schema.ResourceData {
	d := resourceDeveloperSetting().TestResourceData()
	_ = d.Set("virtual_clusters_limit", 2)
	_ = d.Set("cpu", 12)
	_ = d.Set("memory", 16)
	_ = d.Set("storage", 20)
	_ = d.Set("hide_system_cluster_group", false)
	return d
}

func TestResourceDeveloperSettingCRUD(t *testing.T) {
	testResourceCRUD(t, prepareDeveloperSettingResourceData, unitTestMockAPIClient,
		resourceDeveloperSettingCreate, resourceDeveloperSettingRead,
		resourceDeveloperSettingUpdate, resourceDeveloperSettingDelete)
}

func TestResourceDeveloperSettingReadWithoutID(t *testing.T) {
	// Same singleton-guard pattern as password policy: a non-canonical
	// ID should clear the resource rather than flatten stale state.
	d := prepareDeveloperSettingResourceData()
	d.SetId("some-random-id")
	diags := resourceDeveloperSettingRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Empty(t, d.Id(), "Read must clear a non-canonical ID")
}

func TestResourceDeveloperSettingCRUDNegative(t *testing.T) {
	// NOTE: Read's negative path is NOT tested. Both
	// palette-sdk-go/client.GetDeveloperSetting and
	// GetSystemClusterGroupPreference dereference resp.Payload before
	// checking err — same SDK bug pattern as GetPasswordPolicy (see
	// resource_password_policy_test.go). Testing the negative read
	// would SIGSEGV in the SDK, not the provider, so the mock returns
	// success on GET even for the "negative" server.

	t.Run("Create surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Create", prepareDeveloperSettingResourceData,
			unitTestMockAPINegativeClient,
			resourceDeveloperSettingCreate, resourceDeveloperSettingRead,
			resourceDeveloperSettingUpdate, resourceDeveloperSettingDelete,
			false, "Invalid developer credit")
	})

	t.Run("Update surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Update", prepareDeveloperSettingResourceData,
			unitTestMockAPINegativeClient,
			resourceDeveloperSettingCreate, resourceDeveloperSettingRead,
			resourceDeveloperSettingUpdate, resourceDeveloperSettingDelete,
			true, "Invalid developer credit")
	})

	t.Run("Delete surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Delete", prepareDeveloperSettingResourceData,
			unitTestMockAPINegativeClient,
			resourceDeveloperSettingCreate, resourceDeveloperSettingRead,
			resourceDeveloperSettingUpdate, resourceDeveloperSettingDelete,
			true, "Invalid developer credit")
	})
}

// TestToDeveloperSettingOverflow covers the "any value > MaxInt32 →
// fall back to defaults" branch that TestToDeveloperSetting doesn't
// touch. The overflow guard is defensive but reachable in principle
// (schema validation caps at 1000 today, but the ValidateFunc could be
// relaxed later).
func TestToDeveloperSettingOverflow(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()
	// Bypass ValidateFunc by writing directly.
	_ = d.Set("cpu", int(1<<31)) // MaxInt32 + 1
	_ = d.Set("memory", 16)
	_ = d.Set("storage", 20)
	_ = d.Set("virtual_clusters_limit", 2)

	devCredit, sysPref := toDeveloperSetting(d)
	assert.Equal(t, int32(12), devCredit.CPU, "overflow should fall back to default 12")
	assert.Equal(t, int32(16), devCredit.MemoryGiB, "default memory")
	assert.Equal(t, int32(20), devCredit.StorageGiB, "default storage")
	assert.Equal(t, int32(2), devCredit.VirtualClustersLimit, "default virtual clusters limit")
	assert.False(t, sysPref.HideSystemClusterGroups, "overflow branch also resets pref")
}

func TestResourceDeveloperSettingImport(t *testing.T) {
	t.Run("uid matches tenant", func(t *testing.T) {
		d := resourceDeveloperSetting().TestResourceData()
		d.SetId("test-tenant-uid") // matches getMockUserInfoPayload().tenantUid
		got, err := resourceDeveloperSettingImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "default-dev-setting-id", got[0].Id())
	})

	t.Run("uid mismatch errors", func(t *testing.T) {
		d := resourceDeveloperSetting().TestResourceData()
		d.SetId("some-other-tenant-uid")
		_, err := resourceDeveloperSettingImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})
}
