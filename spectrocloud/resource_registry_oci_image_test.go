package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PLT-2484 adds `provider_type = "image"` to spectrocloud_registry_oci (basic
// OCI registries, POST/GET/PUT/DELETE /v1/registries/oci/basic) so Edge
// Native clusters can pull their provider (OS/bootstrap) image from a
// registry that isn't used as a pack source. These tests guard the schema's
// ValidateFunc enum, the toRegistryBasic transform, and the full basic-
// registry CRUD path for that new value.

// TestProviderTypeValidateFunc_AcceptsImage is the direct regression guard
// for the ValidateFunc enum change: before PLT-2484, "image" was rejected
// here with "expected provider_type to be one of [helm zarf pack]".
func TestProviderTypeValidateFunc_AcceptsImage(t *testing.T) {
	validateFunc := resourceRegistryOciEcr().Schema["provider_type"].ValidateFunc
	require.NotNil(t, validateFunc)

	warnings, errs := validateFunc("image", "provider_type")
	assert.Empty(t, warnings)
	assert.Empty(t, errs, "provider_type=image must be accepted by the schema validator")
}

// TestProviderTypeValidateFunc_StillRejectsUnknownValue confirms the enum
// change didn't accidentally open up the field to arbitrary strings.
func TestProviderTypeValidateFunc_StillRejectsUnknownValue(t *testing.T) {
	validateFunc := resourceRegistryOciEcr().Schema["provider_type"].ValidateFunc
	require.NotNil(t, validateFunc)

	_, errs := validateFunc("bogus", "provider_type")
	assert.NotEmpty(t, errs, "an unsupported provider_type value must still be rejected")
}

// prepareImageBasicOciRegistryTestData builds a `type=basic`,
// `provider_type=image` resource, matching the shape of a registry created
// for an Edge Native cluster's image_registry_ref.
//
// credential_type is deliberately "basic" (username/password), not
// "secret"/"sts": toRegistryBasic only maps credential_type=="basic" to a
// basic auth type - every other value (including "secret"/"sts", which are
// ECR-only concepts) silently becomes "noAuth" in the request it builds. An
// earlier version of this fixture used credential_type=secret here, which
// passed but didn't actually exercise the credentials this resource type
// supports.
func prepareImageBasicOciRegistryTestData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-image-registry"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://image-registry.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("provider_type", "image"))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "basic",
			"username":        "image-reg-user",
			"password":        "image-reg-pass",
		},
	}))
	return d
}

// TestToRegistryBasic_ImageProviderType directly checks the request-building
// transform (no mock server involved): provider_type=image must reach the
// SDK spec unchanged, and credential_type=basic must populate Auth.Type and
// username/password - the two things Create/Update actually send over the
// wire.
func TestToRegistryBasic_ImageProviderType(t *testing.T) {
	d := prepareImageBasicOciRegistryTestData(t)
	registry := toRegistryBasic(d)

	require.NotNil(t, registry.Spec.ProviderType)
	assert.Equal(t, "image", *registry.Spec.ProviderType)
	assert.Equal(t, "basic", registry.Spec.Auth.Type)
	assert.Equal(t, "image-reg-user", registry.Spec.Auth.Username)
	assert.Equal(t, "image-reg-pass", registry.Spec.Auth.Password.String())
}

func TestResourceRegistryOci_Image_Create(t *testing.T) {
	d := prepareImageBasicOciRegistryTestData(t)
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.NotEmpty(t, d.Id())
}

// TestResourceRegistryOci_Image_Read exercises Read against a mock fixture
// ("basic-uid-image") whose Spec.ProviderType is explicitly "image", so the
// assertion is tied to that value rather than the mock's unrelated default.
func TestResourceRegistryOci_Image_Read(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "basic"))
	d.SetId("basic-uid-image")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "image", d.Get("provider_type"))
}

func TestResourceRegistryOci_Image_Update(t *testing.T) {
	d := prepareImageBasicOciRegistryTestData(t)
	d.SetId("test-image-registry-uid")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryOci_Image_Delete(t *testing.T) {
	d := prepareImageBasicOciRegistryTestData(t)
	d.SetId("test-image-registry-uid")
	diags := resourceRegistryEcrDelete(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}
