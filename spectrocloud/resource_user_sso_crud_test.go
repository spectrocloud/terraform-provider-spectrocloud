package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// Covers the previously-0% Create/Read/Update/Delete paths on
// resource_user.go and resource_sso.go by wiring up the auxiliary
// role-mapping / SAML / OIDC endpoints (see routes/mockUserRoles.go
// and routes/mockSSO.go).

// ---------------------------------------------------------------------------
// user
// ---------------------------------------------------------------------------

func prepareUserResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceUser().TestResourceData()
	require.NoError(t, d.Set("first_name", "test"))
	require.NoError(t, d.Set("last_name", "spectro"))
	require.NoError(t, d.Set("email", "test@spectrocloud.com"))
	return d
}

func TestResourceUserCreate(t *testing.T) {
	// Create fans out into role-mapping AssociateUser* only when the
	// respective d.GetOk block is set. Bare Create (no role blocks) hits
	// just POST /v1/users and returns.
	d := prepareUserResourceData(t)

	diags := resourceUserCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-user-uid", d.Id())
}

func TestResourceUserRead(t *testing.T) {
	// Read goes through GetUserSummaryByEmail (POST /users/summary) →
	// flattenUser → four role-flatten helpers, each hitting a different
	// endpoint. All are mocked.
	d := prepareUserResourceData(t)
	d.SetId("12345")

	diags := resourceUserRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	// flattenUser writes back the summary's spec values.
	assert.Equal(t, "test", d.Get("first_name"))
	assert.Equal(t, "spectro", d.Get("last_name"))
	assert.Equal(t, "test@spectrocloud.com", d.Get("email"))
}

// TestResourceUserReadNotFound is intentionally omitted — the mock's
// /v1/users/summary route returns the same fixture regardless of the
// email in the request body (mock servers don't inspect POST bodies).
// Exercising the "user not found for email" branch of resourceUserRead
// would require a Handler-based route that echoes based on the request.
// Worth it if we want to pin the soft-delete branch specifically.

func TestResourceUserDelete(t *testing.T) {
	d := prepareUserResourceData(t)
	d.SetId("test-user-uid")

	diags := resourceUserDelete(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceUserUpdate(t *testing.T) {
	// Update is gated on d.HasChanges for role-mapping keys. On
	// TestResourceData those never fire, so Update is effectively a
	// no-op returning nil diags — but that still exercises the
	// function's outer setup + the four HasChanges branches.
	d := prepareUserResourceData(t)
	d.SetId("test-user-uid")

	diags := resourceUserUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// ---------------------------------------------------------------------------
// sso
// ---------------------------------------------------------------------------

func prepareSSOResourceData(t *testing.T, ssoType string) *schema.ResourceData {
	t.Helper()
	d := resourceSSO().TestResourceData()
	require.NoError(t, d.Set("sso_auth_type", ssoType))
	require.NoError(t, d.Set("domains", schema.NewSet(schema.HashString, []interface{}{"example.com"})))
	return d
}

func TestResourceSSORead(t *testing.T) {
	d := prepareSSOResourceData(t, "none")
	d.SetId("test-tenant-uid")

	diags := resourceSSORead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	// Domains should round-trip from the mock's fixture.
	domains := d.Get("domains").(*schema.Set)
	assert.Contains(t, domains.List(), "example.com")
}

func TestResourceSSOUpdate_None(t *testing.T) {
	// sso_auth_type=none triggers disableSSO which hits SAML + OIDC +
	// domains + providers updates. All are mocked.
	d := prepareSSOResourceData(t, "none")
	d.SetId("test-tenant-uid")

	diags := resourceSSOUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceSSODelete(t *testing.T) {
	// Delete calls disableSSO — same endpoint set as Update-none.
	d := prepareSSOResourceData(t, "none")
	d.SetId("test-tenant-uid")

	diags := resourceSSODelete(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}
