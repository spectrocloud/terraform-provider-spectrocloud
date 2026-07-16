package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToResourceLimits(t *testing.T) {
	d := resourceResourceLimit().TestResourceData()

	// Set custom values for testing
	d.Set("alert", 50)
	d.Set("api_keys", 10)
	d.Set("application_deployment", 80)

	resourceLimits, err := toResourceLimits(d)
	assert.NoError(t, err)
	assert.NotNil(t, resourceLimits)

	// Verify specific values
	assert.Equal(t, int64(50), resourceLimits.Resources[0].Limit)
	assert.Equal(t, int64(10), resourceLimits.Resources[1].Limit)
	assert.Equal(t, int64(80), resourceLimits.Resources[2].Limit)
}

func TestToResourceDefaultLimits(t *testing.T) {
	d := resourceResourceLimit().TestResourceData()

	resourceLimits, err := toResourceDefaultLimits(d)
	assert.NoError(t, err)
	assert.NotNil(t, resourceLimits)

	// Verify default values from KindToFieldMapping
	for i, mapping := range KindToFieldMapping {
		assert.Equal(t, mapping.Default, resourceLimits.Resources[i].Limit)
	}
}

func TestFlattenResourceLimits(t *testing.T) {
	d := resourceResourceLimit().TestResourceData()

	resourceLimits := &models.V1TenantResourceLimits{
		Resources: []*models.V1TenantResourceLimit{
			{Kind: models.V1ResourceLimitTypeAlert, Limit: 75},
			{Kind: models.V1ResourceLimitTypeAPIKey, Limit: 15},
			{Kind: models.V1ResourceLimitTypeAppdeployment, Limit: 90},
		},
	}

	err := flattenResourceLimits(resourceLimits, d)
	assert.NoError(t, err)

	// Verify the values were correctly set
	assert.Equal(t, 75, d.Get("alert"))
	assert.Equal(t, 15, d.Get("api_keys"))
	assert.Equal(t, 90, d.Get("application_deployment"))
}

// ---------------------------------------------------------------------------
// CRUD coverage
// ---------------------------------------------------------------------------

func prepareResourceLimitData() *schema.ResourceData {
	d := resourceResourceLimit().TestResourceData()
	// Only set a few — the schema defaults fill in the rest, which
	// keeps this fixture short without making toResourceLimits crash
	// (KindToFieldMapping iterates all 22 fields).
	_ = d.Set("alert", 100)
	_ = d.Set("api_keys", 20)
	_ = d.Set("cluster", 10000)
	return d
}

func TestResourceLimitCRUD(t *testing.T) {
	testResourceCRUD(t, prepareResourceLimitData, unitTestMockAPIClient,
		resourceResourceLimitsCreate, resourceResourceLimitsRead,
		resourceResourceLimitsUpdate, resourceResourceLimitsDelete)
}

func TestResourceLimitReadWithoutID(t *testing.T) {
	// Singleton guard: a non-canonical ID should clear rather than
	// flatten. Same pattern as password_policy / developer_setting.
	d := prepareResourceLimitData()
	d.SetId("some-random-id")
	diags := resourceResourceLimitsRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Empty(t, d.Id())
}

func TestResourceLimitCRUDNegative(t *testing.T) {
	// Read's negative path intentionally not covered — GetResourceLimits
	// derefs resp.Payload before checking err, same SDK bug shared by
	// GetPasswordPolicy / GetDeveloperSetting / GetSystemClusterGroupPreference.

	t.Run("Create surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Create", prepareResourceLimitData,
			unitTestMockAPINegativeClient,
			resourceResourceLimitsCreate, resourceResourceLimitsRead,
			resourceResourceLimitsUpdate, resourceResourceLimitsDelete,
			false, "Invalid resource limits")
	})

	t.Run("Update surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Update", prepareResourceLimitData,
			unitTestMockAPINegativeClient,
			resourceResourceLimitsCreate, resourceResourceLimitsRead,
			resourceResourceLimitsUpdate, resourceResourceLimitsDelete,
			true, "Invalid resource limits")
	})

	t.Run("Delete surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Delete", prepareResourceLimitData,
			unitTestMockAPINegativeClient,
			resourceResourceLimitsCreate, resourceResourceLimitsRead,
			resourceResourceLimitsUpdate, resourceResourceLimitsDelete,
			true, "Invalid resource limits")
	})
}

func TestResourceLimitImport(t *testing.T) {
	t.Run("uid matches tenant", func(t *testing.T) {
		d := resourceResourceLimit().TestResourceData()
		d.SetId("test-tenant-uid")
		got, err := resourceResourceLimitsImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "default-resource-limit-id", got[0].Id())
	})

	t.Run("uid mismatch errors", func(t *testing.T) {
		d := resourceResourceLimit().TestResourceData()
		d.SetId("some-other-tenant-uid")
		_, err := resourceResourceLimitsImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})
}

// TestFlattenResourceLimitsPreservesMissingKinds ensures kinds NOT in the
// payload are left alone (default-preserving behavior). Without this
// guarantee a partial API response would zero out schema fields.
func TestFlattenResourceLimitsPreservesMissingKinds(t *testing.T) {
	d := resourceResourceLimit().TestResourceData()
	_ = d.Set("appliance", 999)

	partial := &models.V1TenantResourceLimits{
		Resources: []*models.V1TenantResourceLimit{
			// Only "alert" — appliance must survive untouched.
			{Kind: models.V1ResourceLimitTypeAlert, Limit: 42},
		},
	}
	require.NoError(t, flattenResourceLimits(partial, d))
	assert.Equal(t, 42, d.Get("alert"))
	assert.Equal(t, 999, d.Get("appliance"),
		"kind not in payload should not overwrite existing schema value")
}
