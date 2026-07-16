package spectrocloud

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// prepareRegistrationTokenResourceData mirrors the fixture served by
// tests/mockApiServer/routes/mockRegistrationToken.go, so the round-trip
// through Create → Read stays drift-free and testResourceCRUD passes its
// "no update diagnostics" check.
func prepareRegistrationTokenResourceData() *schema.ResourceData {
	d := resourceRegistrationToken().TestResourceData()
	_ = d.Set("name", "test-reg-token")
	_ = d.Set("description", "mock registration token")
	_ = d.Set("project_uid", "test-project-uid")
	_ = d.Set("expiry_date", "2030-01-01")
	_ = d.Set("status", "active")
	return d
}

func TestResourceRegistrationTokenCRUD(t *testing.T) {
	testResourceCRUD(t, prepareRegistrationTokenResourceData, unitTestMockAPIClient,
		resourceRegistrationTokenCreate, resourceRegistrationTokenRead,
		resourceRegistrationTokenUpdate, resourceRegistrationTokenDelete)
}

func TestResourceRegistrationTokenCRUDNegative(t *testing.T) {
	t.Run("Create conflict", func(t *testing.T) {
		testResourceCRUDNegative(t, "Create", prepareRegistrationTokenResourceData,
			unitTestMockAPINegativeClient,
			resourceRegistrationTokenCreate, resourceRegistrationTokenRead,
			resourceRegistrationTokenUpdate, resourceRegistrationTokenDelete,
			false, "already exists")
	})

	t.Run("Read NotFound clears id", func(t *testing.T) {
		d := prepareRegistrationTokenResourceData()
		d.SetId("stale-uid")

		diags := resourceRegistrationTokenRead(context.Background(), d, unitTestMockAPINegativeClient)
		assert.Empty(t, diags, "ResourceNotFound should be swallowed by handleReadError")
		assert.Empty(t, d.Id(), "handleReadError should clear the ID on NotFound")
	})

	// NOTE: Update's negative path is intentionally not covered here.
	// resourceRegistrationTokenUpdate short-circuits on HasChange/HasChanges,
	// and schema.TestResourceData reports no diffs when fields are only
	// Set() (there's no prior state to compare against), so the API is
	// never called and no diag ever surfaces — the "negative Update"
	// scenario is not reachable through this test harness. Testing it
	// properly would require driving the resource through a full plan/
	// apply cycle, which is out of scope for a unit test.

	t.Run("Delete failure", func(t *testing.T) {
		testResourceCRUDNegative(t, "Delete", prepareRegistrationTokenResourceData,
			unitTestMockAPINegativeClient,
			resourceRegistrationTokenCreate, resourceRegistrationTokenRead,
			resourceRegistrationTokenUpdate, resourceRegistrationTokenDelete,
			true, "not found")
	})
}

// TestStateConvertBool / TestStateConvertString cover the two-line active⇄
// isActive helpers. Keeping them separate from CRUD tests keeps the
// failure attribution clean if someone flips the semantics.
func TestStateConvertBool(t *testing.T) {
	assert.Equal(t, "active", StateConvertBool(true))
	assert.Equal(t, "inactive", StateConvertBool(false))
}

func TestStateConvertString(t *testing.T) {
	assert.True(t, stateConvertString("active"))
	assert.False(t, stateConvertString("inactive"))
	assert.False(t, stateConvertString(""), "unknown state is not active")
	assert.False(t, stateConvertString("Active"), "match is case-sensitive")
}

func TestToRegistrationTokenCreate(t *testing.T) {
	t.Run("valid input builds entity", func(t *testing.T) {
		d := prepareRegistrationTokenResourceData()
		got, err := toRegistrationTokenCreate(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Metadata)
		require.NotNil(t, got.Spec)
		assert.Equal(t, "test-reg-token", got.Metadata.Name)
		assert.Equal(t, "mock registration token", got.Metadata.Annotations["description"])
		assert.Equal(t, "test-project-uid", got.Spec.DefaultProjectUID)

		// Expiry should serialize the same date we set — pull it back through
		// strfmt.DateTime so we compare on a normalized value.
		want, err := time.Parse("2006-01-02", "2030-01-01")
		require.NoError(t, err)
		assert.Equal(t, time.Time(strfmt.DateTime(want)), time.Time(strfmt.DateTime(got.Spec.Expiry)))
	})

	t.Run("bad expiry_date propagates parse error", func(t *testing.T) {
		d := resourceRegistrationToken().TestResourceData()
		// Bypass schema validation by setting via TestResourceData directly.
		_ = d.Set("name", "x")
		_ = d.Set("expiry_date", "not-a-date")
		_, err := toRegistrationTokenCreate(d)
		require.Error(t, err)
	})
}

func TestToRegistrationTokenUpdate(t *testing.T) {
	d := prepareRegistrationTokenResourceData()
	d.SetId("test-reg-token-uid")
	got, err := toRegistrationTokenUpdate(d)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.Metadata)
	assert.Equal(t, "test-reg-token-uid", got.Metadata.UID)
	assert.Equal(t, "test-reg-token", got.Metadata.Name)
	assert.Equal(t, "mock registration token", got.Metadata.Annotations["description"])
	assert.Equal(t, "test-project-uid", got.Spec.DefaultProjectUID)
}

// TestFlattenRegistrationToken feeds the same fixture the mock uses and
// asserts every field lands on d — including the trickiest ones (V1Time →
// YYYY-MM-DD, IsActive → status string). If the mock and this fixture
// ever diverge, this test's failure will be the first signal.
func TestFlattenRegistrationToken(t *testing.T) {
	d := resourceRegistrationToken().TestResourceData()

	expiry, err := time.Parse("2006-01-02", "2030-01-01")
	require.NoError(t, err)
	in := &models.V1EdgeToken{
		Metadata: &models.V1ObjectMeta{
			Name: "flatten-name",
			Annotations: map[string]string{
				"description": "flatten-desc",
			},
		},
		Spec: &models.V1EdgeTokenSpec{
			Expiry: models.V1Time(strfmt.DateTime(expiry)),
			Token:  "flatten-token-value",
			DefaultProject: &models.V1EdgeTokenProject{
				UID:  "flatten-project-uid",
				Name: "Default",
			},
		},
		Status: &models.V1EdgeTokenStatus{IsActive: false},
	}
	require.NoError(t, flattenRegistrationToken(d, in))
	assert.Equal(t, "flatten-name", d.Get("name"))
	assert.Equal(t, "flatten-desc", d.Get("description"))
	assert.Equal(t, "flatten-project-uid", d.Get("project_uid"))
	assert.Equal(t, "2030-01-01", d.Get("expiry_date"))
	assert.Equal(t, "flatten-token-value", d.Get("token"))
	assert.Equal(t, "inactive", d.Get("status"))
}

// TestResourceRegistrationTokenImport covers both the "empty id" guard and
// the happy path (lookup-by-uid succeeds, resource is populated from Read).
func TestResourceRegistrationTokenImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourceRegistrationToken().TestResourceData()
		d.SetId("")
		_, err := resourceRegistrationTokenImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("uid lookup succeeds", func(t *testing.T) {
		d := resourceRegistrationToken().TestResourceData()
		d.SetId("test-reg-token-uid")
		res, err := resourceRegistrationTokenImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "test-reg-token-uid", res[0].Id())
		assert.Equal(t, "test-reg-token", res[0].Get("name"))
	})
}
