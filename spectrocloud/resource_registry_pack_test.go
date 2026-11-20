package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// PLT-2441: spectrocloud_registry_pack manages Pack registries through /v1/registries/pack.

func packRegistryCredentials(credentialType string, extra map[string]interface{}) []interface{} {
	cred := map[string]interface{}{"credential_type": credentialType}
	for k, v := range extra {
		cred[k] = v
	}
	return []interface{}{cred}
}

func prepareResourceRegistryPack() *schema.ResourceData {
	d := resourceRegistryPack().TestResourceData()
	_ = d.Set("name", "test-pack-registry")
	_ = d.Set("endpoint", "https://pack.example.com:5000")
	_ = d.Set("credentials", packRegistryCredentials("basic", map[string]interface{}{
		"username": "pack-user",
		"password": "pack-password",
	}))
	return d
}

func packRegistryCredentialFromState(t *testing.T, d *schema.ResourceData) map[string]interface{} {
	t.Helper()
	creds := d.Get("credentials").([]interface{})
	require.Len(t, creds, 1)
	return creds[0].(map[string]interface{})
}

func TestResourceRegistryPackSchema(t *testing.T) {
	p := New("test")()
	require.Contains(t, p.ResourcesMap, "spectrocloud_registry_pack")
	require.NoError(t, p.InternalValidate())

	s := resourceRegistryPack().Schema
	assert.True(t, s["name"].ForceNew, "a Pack registry cannot be renamed in place")
	assert.False(t, s["endpoint"].ForceNew, "the endpoint is updated in place")
	cred := s["credentials"].Elem.(*schema.Resource).Schema
	assert.True(t, cred["password"].Sensitive)
	assert.True(t, cred["username"].Required)
	assert.True(t, cred["password"].Required)
	assert.NotContains(t, cred, "token", "Pack registries do not support token authentication")
	assert.True(t, cred["tls_config"].Elem.(*schema.Resource).Schema["key"].Sensitive)
	assert.Equal(t, "basic", cred["credential_type"].Default)
}

// Pack registries only support basic (username and password) authentication.
func TestResourceRegistryPackCredentialTypeValidation(t *testing.T) {
	validate := resourceRegistryPack().Schema["credentials"].Elem.(*schema.Resource).Schema["credential_type"].ValidateFunc
	_, errs := validate("basic", "credential_type")
	assert.Empty(t, errs)
	for _, unsupported := range []string{"noAuth", "token"} {
		_, errs = validate(unsupported, "credential_type")
		assert.NotEmpty(t, errs, "%s must be rejected at plan time", unsupported)
	}
}

func TestResourceRegistryPackCRUD(t *testing.T) {
	testResourceCRUD(t, prepareResourceRegistryPack, unitTestMockAPIClient,
		resourceRegistryPackCreate, resourceRegistryPackRead, resourceRegistryPackUpdate, resourceRegistryPackDelete)
}

func TestResourceRegistryPackCreate(t *testing.T) {
	d := prepareResourceRegistryPack()
	diags := resourceRegistryPackCreate(context.Background(), d, unitTestMockAPIClient)
	require.Empty(t, diags)
	assert.Equal(t, routes.PackRegistryUID, d.Id())
	assert.Equal(t, "https://pack.example.com:5000", d.Get("endpoint"))

	cred := packRegistryCredentialFromState(t, d)
	assert.Equal(t, "basic", cred["credential_type"])
	assert.Equal(t, "pack-user", cred["username"])
	assert.Equal(t, "pack-password", cred["password"], "Palette does not return the password; the configured value must stay")
}

func TestResourceRegistryPackCreateFailsValidation(t *testing.T) {
	d := prepareResourceRegistryPack()
	_ = d.Set("endpoint", routes.PackRegistryUnreachable)

	diags := resourceRegistryPackCreate(context.Background(), d, unitTestMockAPIClient)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, `pack registry "test-pack-registry" failed validation`)
	assert.Contains(t, diags[0].Summary, "unable to reach the pack registry")
	assert.Empty(t, d.Id(), "a registry that fails validation must not be created")
}

func TestToRegistryPack(t *testing.T) {
	d := prepareResourceRegistryPack()
	_ = d.Set("credentials", packRegistryCredentials("basic", map[string]interface{}{
		"username": "pack-user",
		"password": "pack-password",
		"tls_config": []interface{}{map[string]interface{}{
			"enabled":              true,
			"ca":                   "ca-pem",
			"certificate":          "cert-pem",
			"key":                  "key-pem",
			"insecure_skip_verify": false,
		}},
	}))

	registry := toRegistryPack(d)
	require.NotNil(t, registry.Spec)
	assert.Equal(t, "test-pack-registry", registry.Metadata.Name)
	assert.Equal(t, "test-pack-registry", registry.Spec.Name)
	assert.Equal(t, "https://pack.example.com:5000", *registry.Spec.Endpoint)
	assert.False(t, registry.Spec.Private, "the resource does not expose private; the API default false is sent")
	assert.Equal(t, "basic", registry.Spec.Auth.Type)
	assert.Equal(t, "pack-user", registry.Spec.Auth.Username)
	assert.Equal(t, "pack-password", registry.Spec.Auth.Password.String())
	require.NotNil(t, registry.Spec.Auth.TLS)
	assert.Equal(t, "ca-pem", registry.Spec.Auth.TLS.Ca)
	assert.Equal(t, "key-pem", registry.Spec.Auth.TLS.Key)

	_ = d.Set("credentials", packRegistryCredentials("basic", map[string]interface{}{
		"username": "pack-user",
		"password": "pack-password",
	}))
	registry = toRegistryPack(d)
	assert.Equal(t, "basic", registry.Spec.Auth.Type)
	assert.Nil(t, registry.Spec.Auth.TLS, "no tls_config block means no TLS configuration is sent")
}

func TestResourceRegistryPackRead(t *testing.T) {
	t.Run("basic reads the username and keeps the configured password", func(t *testing.T) {
		d := prepareResourceRegistryPack()
		d.SetId(routes.PackRegistryUID)

		require.Empty(t, resourceRegistryPackRead(context.Background(), d, unitTestMockAPIClient))
		cred := packRegistryCredentialFromState(t, d)
		assert.Equal(t, "basic", cred["credential_type"])
		assert.Equal(t, "pack-user", cred["username"])
		assert.Equal(t, "pack-password", cred["password"], "Palette does not return the password; the configured value must stay")
		assert.Empty(t, cred["tls_config"], "no TLS configured and none returned")
	})

	t.Run("tls configuration is read back", func(t *testing.T) {
		d := prepareResourceRegistryPack()
		d.SetId(routes.PackRegistryTLSUID)

		require.Empty(t, resourceRegistryPackRead(context.Background(), d, unitTestMockAPIClient))
		tls := packRegistryCredentialFromState(t, d)["tls_config"].([]interface{})
		require.Len(t, tls, 1)
		assert.Equal(t, "test-ca-pem", tls[0].(map[string]interface{})["ca"])
		assert.Equal(t, true, tls[0].(map[string]interface{})["insecure_skip_verify"])
	})

	t.Run("deleted outside Terraform clears the ID", func(t *testing.T) {
		d := prepareResourceRegistryPack()
		d.SetId(routes.PackRegistryMissingUID)

		require.Empty(t, resourceRegistryPackRead(context.Background(), d, unitTestMockAPIClient))
		assert.Empty(t, d.Id())
	})
}

func TestResourceRegistryPackDeleteAlreadyGone(t *testing.T) {
	d := prepareResourceRegistryPack()
	d.SetId(routes.PackRegistryMissingUID)
	assert.Empty(t, resourceRegistryPackDelete(context.Background(), d, unitTestMockAPIClient))
}

// packRegistryUpdateData builds the ResourceData Terraform passes to Update when only the endpoint changes, so
// HasChanges sees the change the way it does during a real apply.
func packRegistryUpdateData(t *testing.T, oldEndpoint, newEndpoint string) *schema.ResourceData {
	t.Helper()
	res := resourceRegistryPack()
	raw := func(endpoint string) map[string]interface{} {
		return map[string]interface{}{
			"name":        "test-pack-registry",
			"endpoint":    endpoint,
			"credentials": packRegistryCredentials("basic", map[string]interface{}{"username": "pack-user", "password": "pack-password"}),
		}
	}
	old := schema.TestResourceDataRaw(t, res.Schema, raw(oldEndpoint))
	old.SetId(routes.PackRegistryUID)
	state := old.State()
	diff, err := res.Diff(context.Background(), state, terraform.NewResourceConfigRaw(raw(newEndpoint)), nil)
	require.NoError(t, err)
	d, err := schema.InternalMap(res.Schema).Data(state, diff)
	require.NoError(t, err)
	d.SetId(routes.PackRegistryUID)
	return d
}

func TestResourceRegistryPackUpdate(t *testing.T) {
	t.Run("endpoint change is validated", func(t *testing.T) {
		d := packRegistryUpdateData(t, "https://pack.example.com:5000", routes.PackRegistryUnreachable)
		diags := resourceRegistryPackUpdate(context.Background(), d, unitTestMockAPIClient)
		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary, "failed validation")
	})

	t.Run("endpoint change that passes validation is applied", func(t *testing.T) {
		d := packRegistryUpdateData(t, "https://old.pack.example.com", "https://pack.example.com:5000")
		assert.Empty(t, resourceRegistryPackUpdate(context.Background(), d, unitTestMockAPIClient))
	})

	t.Run("wait_for_sync surfaces a failed sync", func(t *testing.T) {
		d := prepareResourceRegistryPack()
		d.SetId(routes.PackRegistrySyncFailUID)
		_ = d.Set("wait_for_sync", true)

		diags := resourceRegistryPackUpdate(context.Background(), d, unitTestMockAPIClient)
		require.True(t, diags.HasError())
		assert.Equal(t, "Pack registry sync failed", diags[0].Summary)
		assert.Contains(t, diags[0].Detail, "index.yaml not found")
	})
}

func TestResourceRegistryPackCreateWithWaitForSync(t *testing.T) {
	d := prepareResourceRegistryPack()
	_ = d.Set("wait_for_sync", true)
	assert.Empty(t, resourceRegistryPackCreate(context.Background(), d, unitTestMockAPIClient))
	assert.Equal(t, routes.PackRegistryUID, d.Id())
}

func TestResourceRegistryPackImport(t *testing.T) {
	t.Run("by UID", func(t *testing.T) {
		d := resourceRegistryPack().TestResourceData()
		d.SetId(routes.PackRegistryUID)
		out, err := resourceRegistryPackImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, out, 1)
		assert.Equal(t, routes.PackRegistryUID, out[0].Id())
		assert.Equal(t, "test-pack-registry", out[0].Get("name"))
		assert.Equal(t, "basic", packRegistryCredentialFromState(t, out[0])["credential_type"])
	})

	t.Run("by name", func(t *testing.T) {
		d := resourceRegistryPack().TestResourceData()
		d.SetId("test-registry-name") // name of the registry returned by GET /v1/registries/pack
		out, err := resourceRegistryPackImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		assert.Equal(t, "test-registry-uid", out[0].Id())
	})

	t.Run("unknown", func(t *testing.T) {
		d := resourceRegistryPack().TestResourceData()
		d.SetId(routes.PackRegistryMissingUID)
		_, err := resourceRegistryPackImport(context.Background(), d, unitTestMockAPIClient)
		assert.Error(t, err)
	})
}
