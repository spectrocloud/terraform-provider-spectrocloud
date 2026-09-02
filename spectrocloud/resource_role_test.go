package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToRole(t *testing.T) {
	d := resourceRole().TestResourceData()
	err := d.Set("name", "test-role")
	if err != nil {
		return
	}
	err = d.Set("type", "project")
	if err != nil {
		return
	}
	err = d.Set("permissions", []interface{}{"bbb"})
	if err != nil {
		return
	}

	role := toRole(d)

	expected := &models.V1Role{
		Metadata: &models.V1ObjectMeta{
			Annotations: map[string]string{
				"scope": "project",
			},
			Name: "test-role",
		},
		Spec: &models.V1RoleSpec{
			Permissions: []string{"bbb"},
			Scope:       models.V1Scope("project"),
			Type:        "user",
		},
		Status: &models.V1RoleStatus{
			IsEnabled: true,
		},
	}

	assert.Equal(t, expected, role)
}

func TestFlattenRole(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"type": {
			Type:     schema.TypeString,
			Required: true,
		},
		"permissions": {
			Type:     schema.TypeSet,
			Required: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}, map[string]interface{}{})

	role := &models.V1Role{
		Metadata: &models.V1ObjectMeta{
			Name: "test-role",
		},
		Spec: &models.V1RoleSpec{
			Permissions: []string{"read", "write"},
			Scope:       models.V1Scope("admin"),
		},
	}

	err := flattenRole(d, role)
	assert.NoError(t, err)
	assert.Equal(t, "test-role", d.Get("name"))
	assert.Equal(t, "admin", d.Get("type"))
	assert.ElementsMatch(t, []interface{}{"read", "write"}, d.Get("permissions").(*schema.Set).List())
}

// ---------------------------------------------------------------------------
// CRUD coverage
// ---------------------------------------------------------------------------

func prepareRoleResourceData() *schema.ResourceData {
	d := resourceRole().TestResourceData()
	_ = d.Set("name", "test-role")
	_ = d.Set("type", "project")
	_ = d.Set("permissions", []interface{}{"perm1", "perm2"})
	return d
}

func TestResourceRoleCRUD(t *testing.T) {
	testResourceCRUD(t, prepareRoleResourceData, unitTestMockAPIClient,
		resourceRoleCreate, resourceRoleRead, resourceRoleUpdate, resourceRoleDelete)
}

func TestResourceRoleCRUDNegative(t *testing.T) {
	t.Run("Create conflict", func(t *testing.T) {
		testResourceCRUDNegative(t, "Create", prepareRoleResourceData,
			unitTestMockAPINegativeClient,
			resourceRoleCreate, resourceRoleRead, resourceRoleUpdate, resourceRoleDelete,
			false, "already exists")
	})

	t.Run("Read NotFound clears id", func(t *testing.T) {
		d := prepareRoleResourceData()
		d.SetId("stale-uid")
		diags := resourceRoleRead(context.Background(), d, unitTestMockAPINegativeClient)
		assert.Empty(t, diags, "ResourceNotFound should be swallowed by handleReadError")
		assert.Empty(t, d.Id(), "handleReadError should clear the ID on NotFound")
	})

	t.Run("Update failure", func(t *testing.T) {
		testResourceCRUDNegative(t, "Update", prepareRoleResourceData,
			unitTestMockAPINegativeClient,
			resourceRoleCreate, resourceRoleRead, resourceRoleUpdate, resourceRoleDelete,
			true, "Invalid role update")
	})

	t.Run("Delete failure", func(t *testing.T) {
		testResourceCRUDNegative(t, "Delete", prepareRoleResourceData,
			unitTestMockAPINegativeClient,
			resourceRoleCreate, resourceRoleRead, resourceRoleUpdate, resourceRoleDelete,
			true, "not found")
	})
}

// TestConvertInterfaceSliceToStringSlice exercises the util's happy path
// and its non-string branch (the second return value). Kept as its own
// test since the function's sole caller — toRole — silently discards
// the error today (see resource_role.go line 154), so any regression
// there wouldn't show up in CRUD tests.
func TestConvertInterfaceSliceToStringSlice(t *testing.T) {
	t.Run("all strings", func(t *testing.T) {
		got, err := convertInterfaceSliceToStringSlice([]interface{}{"a", "b", "c"})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, got)
	})

	t.Run("non-string errors", func(t *testing.T) {
		_, err := convertInterfaceSliceToStringSlice([]interface{}{"a", 42})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a string")
	})

	t.Run("empty input returns nil slice", func(t *testing.T) {
		got, err := convertInterfaceSliceToStringSlice([]interface{}{})
		require.NoError(t, err)
		assert.Nil(t, got, "no elements → nil, per current append-based implementation")
	})
}

func TestResourceRoleImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourceRole().TestResourceData()
		d.SetId("")
		_, err := resourceRoleImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("uid lookup succeeds", func(t *testing.T) {
		d := resourceRole().TestResourceData()
		d.SetId(mockRoleUIDForTest)
		got, err := resourceRoleImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, mockRoleUIDForTest, got[0].Id())
		assert.Equal(t, "test-role", got[0].Get("name"))
	})
}

// mockRoleUIDForTest mirrors the constant defined in
// tests/mockApiServer/routes/mockRole.go. Duplicating it here (rather
// than reaching across the package boundary) keeps this test file
// self-contained; if the mock's UID changes, the assertion here will
// visibly break in the same PR.
const mockRoleUIDForTest = "test-role-uid"
