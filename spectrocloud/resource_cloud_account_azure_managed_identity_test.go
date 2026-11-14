package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// PLT-2489 adds UserAssignedMSI support to spectrocloud_cloudaccount_azure:
// a new `identity_type` field (ServicePrincipal default / UserAssignedMSI),
// `azure_client_secret` becoming conditionally required, and a computed
// `management_identity_resource_id` populated from the API.

// TestIdentityType_ForceNew guards the "no conversion" product constraint:
// PCP-3804 puts converting an existing service-principal account to managed
// identity authentication out of scope, so changing identity_type must
// replace the resource rather than update it in place.
func TestIdentityType_ForceNew(t *testing.T) {
	fieldSchema, ok := resourceCloudAccountAzure().Schema["identity_type"]
	require.True(t, ok, "resourceCloudAccountAzure must define identity_type")
	assert.True(t, fieldSchema.ForceNew)
	assert.Equal(t, "ServicePrincipal", fieldSchema.Default)
}

// TestManagementIdentityResourceID_Computed guards that the field is
// read-only: Palette resolves it during validation, Terraform never sends it.
func TestManagementIdentityResourceID_Computed(t *testing.T) {
	fieldSchema, ok := resourceCloudAccountAzure().Schema["management_identity_resource_id"]
	require.True(t, ok)
	assert.True(t, fieldSchema.Computed)
	assert.False(t, fieldSchema.Required)
	assert.False(t, fieldSchema.Optional)
}

// TestValidateAzureAccountIdentityType is the direct regression guard for
// the CustomizeDiff body: ServicePrincipal (including the unset/default
// case) requires azure_client_secret; UserAssignedMSI does not.
func TestValidateAzureAccountIdentityType(t *testing.T) {
	tests := []struct {
		name         string
		identityType string
		clientSecret string
		wantErr      bool
	}{
		{"unset identity_type with secret", "", "secret", false},
		{"unset identity_type without secret", "", "", true},
		{"ServicePrincipal with secret", "ServicePrincipal", "secret", false},
		{"ServicePrincipal without secret", "ServicePrincipal", "", true},
		{"UserAssignedMSI without secret", "UserAssignedMSI", "", false},
		{"UserAssignedMSI with secret present anyway", "UserAssignedMSI", "secret", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAzureAccountIdentityType(tt.identityType, tt.clientSecret)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestToAzureAccount_UserAssignedMSI_OmitsClientSecret drives toAzureAccount
// directly: a UserAssignedMSI account with no azure_client_secret must send
// a nil ClientSecret pointer (omitted from the request), not a pointer to an
// empty string, and must carry IdentityType through.
func TestToAzureAccount_UserAssignedMSI_OmitsClientSecret(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	require.NoError(t, rd.Set("name", "test-msi-account"))
	require.NoError(t, rd.Set("context", "project"))
	require.NoError(t, rd.Set("azure_tenant_id", "tenant-id"))
	require.NoError(t, rd.Set("azure_client_id", "msi-client-id"))
	require.NoError(t, rd.Set("identity_type", "UserAssignedMSI"))
	require.NoError(t, rd.Set("private_cloud_gateway_id", "test-pcg-id"))

	account := toAzureAccount(rd)
	require.NotNil(t, account.Spec.IdentityType)
	assert.Equal(t, "UserAssignedMSI", *account.Spec.IdentityType)
	assert.Nil(t, account.Spec.ClientSecret, "ClientSecret must be omitted (nil), not an empty-string pointer")
}

// TestToAzureAccount_ServicePrincipal_SetsClientSecret confirms the
// pre-existing behavior for the default identity type is unchanged.
func TestToAzureAccount_ServicePrincipal_SetsClientSecret(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	require.NoError(t, rd.Set("name", "test-sp-account"))
	require.NoError(t, rd.Set("context", "project"))
	require.NoError(t, rd.Set("azure_tenant_id", "tenant-id"))
	require.NoError(t, rd.Set("azure_client_id", "sp-client-id"))
	require.NoError(t, rd.Set("azure_client_secret", "sp-secret"))
	require.NoError(t, rd.Set("identity_type", "ServicePrincipal"))

	account := toAzureAccount(rd)
	require.NotNil(t, account.Spec.ClientSecret)
	assert.Equal(t, "sp-secret", *account.Spec.ClientSecret)
}

// TestFlattenCloudAccountAzure_IdentityType covers the Read-side default:
// an API response with no IdentityType set must flatten to "ServicePrincipal"
// (pre-PLT-2489 accounts), not an empty string.
func TestFlattenCloudAccountAzure_IdentityType_DefaultsWhenUnset(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	account := &models.V1AzureAccount{
		Metadata: &models.V1ObjectMeta{Name: "acc", Annotations: map[string]string{}},
		Spec: &models.V1AzureCloudAccount{
			ClientID:   strPtrAzureTest("client-id"),
			TenantID:   strPtrAzureTest("tenant-id"),
			TenantName: "tenant",
			Settings:   &models.V1CloudAccountSettings{},
		},
	}
	diags, hasError := flattenCloudAccountAzure(rd, account)
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "ServicePrincipal", rd.Get("identity_type"))
	assert.Equal(t, "", rd.Get("management_identity_resource_id"))
}

// TestFlattenCloudAccountAzure_IdentityType_MSI covers the populated case.
func TestFlattenCloudAccountAzure_IdentityType_MSI(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	identityType := models.V1AzureCloudAccountIdentityTypeUserAssignedMSI
	account := &models.V1AzureAccount{
		Metadata: &models.V1ObjectMeta{Name: "acc", Annotations: map[string]string{}},
		Spec: &models.V1AzureCloudAccount{
			ClientID:                     strPtrAzureTest("msi-client-id"),
			TenantID:                     strPtrAzureTest("tenant-id"),
			TenantName:                   "tenant",
			IdentityType:                 &identityType,
			ManagementIdentityResourceID: "resolved-mi-resource-id",
			Settings:                     &models.V1CloudAccountSettings{},
		},
	}
	diags, hasError := flattenCloudAccountAzure(rd, account)
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "UserAssignedMSI", rd.Get("identity_type"))
	assert.Equal(t, "resolved-mi-resource-id", rd.Get("management_identity_resource_id"))
}

func strPtrAzureTest(s string) *string { return &s }

// TestResourceCloudAccountAzureRead_MSI drives the full Read path against
// the dedicated mock fixture (routes.AzureMSIAccountUID), proving the
// end-to-end wiring, not just the pure flatten function.
func TestResourceCloudAccountAzureRead_MSI(t *testing.T) {
	d := resourceCloudAccountAzure().TestResourceData()
	d.SetId(routes.AzureMSIAccountUID)
	require.NoError(t, d.Set("context", "project"))

	diags := resourceCloudAccountAzureRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "UserAssignedMSI", d.Get("identity_type"))
	assert.Equal(t, "test-managed-identity-resource-id", d.Get("management_identity_resource_id"))
}

// TestResourceCloudAccountAzureCreate_MSI drives Create end-to-end for a
// UserAssignedMSI account (no azure_client_secret), exercising the new
// overlord-scoped validate branch in validateCloudAccountAzure (SDK) via the
// mock's POST /v1/overlords/azure/{uid}/account/validate route.
func TestResourceCloudAccountAzureCreate_MSI(t *testing.T) {
	d := resourceCloudAccountAzure().TestResourceData()
	require.NoError(t, d.Set("name", "test-msi-account"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("azure_tenant_id", "tenant-id"))
	require.NoError(t, d.Set("azure_client_id", "msi-client-id"))
	require.NoError(t, d.Set("identity_type", "UserAssignedMSI"))
	require.NoError(t, d.Set("private_cloud_gateway_id", "test-pcg-id"))
	require.NoError(t, d.Set("cloud", "AzurePublicCloud"))

	diags := resourceCloudAccountAzureCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.NotEmpty(t, d.Id())
}
