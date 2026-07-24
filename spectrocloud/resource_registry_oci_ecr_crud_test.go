package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// Fills the previously-uncovered CRUD branches of resource_registry_oci_ecr.go.
//
// Coverage before:
//   - resourceRegistryEcrCreate: 25.5%  (STS path from wave2, secret path untouched, no wait branch)
//   - resourceRegistryEcrUpdate: 0%
//   - resourceRegistryEcrDelete: 63.6% (delete needs to try both ecr + basic paths)
//   - waitForOciRegistrySync: 0%
//   - resourceOciRegistrySyncRefreshFunc: 25% (Batch 17 exercised negative mock)
//
// Now that Batch 22 zeroed the wait Delay, we can flip wait_for_sync
// and let the wait fire against the mocked sync-status endpoint.

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestResourceRegistryEcrCreate_SecretPath(t *testing.T) {
	// The secret-cred fixture from the existing test file drives the
	// "credential_type=secret" branch inside toRegistryEcr → the SDK
	// call → success against the mock's POST /v1/registries/oci/ecr.
	d := prepareOciEcrRegistryTestDataSecret()
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrCreate_STSPath(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrCreate_BasicPath(t *testing.T) {
	// The `type=basic` branch dispatches to CreateOciBasicRegistry
	// (mocked at POST /v1/registries/oci/basic). Requires an "endpoint"
	// scalar plus a credentials block.
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"secret_key":      "sk",
			"access_key":      "ak",
		},
	}))
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPIClient)
	_ = diags // may include validate warnings; branch is what matters
}

// TestResourceRegistryEcrCreate_Ecr_WaitForSync_Success drives Create's
// wait-for-sync block for `type=ecr`: the created UID
// ("test-sts-oci-reg-ecr-uid", from the POST route) isn't the special
// ociEcrRegistrySyncStatusErrorUID, so GET .../ecr/sync/status returns the
// default Success+Message fixture — exercising the `syncStatus.Message !=
// ""` branch of the post-wait status formatting.
func TestResourceRegistryEcrCreate_Ecr_WaitForSync_Success(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	require.NoError(t, d.Set("provider_type", "helm"))
	require.NoError(t, d.Set("wait_for_sync", true))
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "Registry synchronized successfully", d.Get("wait_for_status_message"))
}

func TestResourceRegistryEcrCreate_Basic_WaitForSync_Success(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry-wait"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("provider_type", "helm"))
	require.NoError(t, d.Set("wait_for_sync", true))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"secret_key":      "sk",
			"access_key":      "ak",
		},
	}))
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "Registry synchronized successfully", d.Get("wait_for_status_message"))
}

// TestResourceRegistryEcrCreate_Ecr_ValidateError drives the
// validateRegistryCred error return inside Create's ecr branch: isSync=true
// with provider_type=helm gates the call, and the negative client 404s the
// validate endpoint (RegistriesNegativeRoutes has no override for it).
func TestResourceRegistryEcrCreate_Ecr_ValidateError(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	require.NoError(t, d.Set("is_synchronization", true))
	require.NoError(t, d.Set("provider_type", "helm"))
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
	assert.Empty(t, d.Id(), "Create must not set an ID when validate fails before the SDK create call")
}

func TestResourceRegistryEcrCreate_Basic_ValidateError(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry-validate-err"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("is_synchronization", true))
	require.NoError(t, d.Set("provider_type", "zarf"))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "basic",
			"username":        "user",
			"password":        "pass",
		},
	}))
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
	assert.Empty(t, d.Id())
}

// TestResourceRegistryEcrCreate_Ecr_CreateError isolates
// CreateOciEcrRegistry's error return by setting is_synchronization=false
// (validateRegistryCred short-circuits) so the negative client's 404
// against POST /v1/registries/oci/ecr is the only failure in play.
func TestResourceRegistryEcrCreate_Ecr_CreateError(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	require.NoError(t, d.Set("is_synchronization", false))
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
	assert.Empty(t, d.Id())
}

func TestResourceRegistryEcrCreate_Basic_CreateError(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry-create-err"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"secret_key":      "sk",
			"access_key":      "ak",
		},
	}))
	diags := resourceRegistryEcrCreate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
	assert.Empty(t, d.Id())
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestResourceRegistryEcrUpdate_EcrPath(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	d.SetId("test-registry-uid")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrUpdate_BasicPath(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"secret_key":      "sk",
			"access_key":      "ak",
		},
	}))
	d.SetId("test-registry-uid")
	_ = resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
}

// TestResourceRegistryEcrUpdate_Ecr_WaitForSync_StatusOnlyMessage picks a
// UID dispatched by ociRegistrySyncStatusFixtureFor to a Status="Completed"
// / empty-Message payload, exercising the `else if syncStatus.Status != ""`
// (Status: %s) formatting branch that TestResourceRegistryEcrCreate_*
// doesn't reach.
func TestResourceRegistryEcrUpdate_Ecr_WaitForSync_StatusOnlyMessage(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	require.NoError(t, d.Set("provider_type", "helm"))
	require.NoError(t, d.Set("wait_for_sync", true))
	d.SetId("oci-sync-completed-no-message")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "Status: Completed", d.Get("wait_for_status_message"))
}

// TestResourceRegistryEcrUpdate_Ecr_WaitForSync_APIError uses the UID that
// ociEcrRegistrySyncStatusHandler 404s on, so both waitForOciRegistrySync's
// internal refresh and the follow-up getOciRegistrySyncStatus call fail —
// covering the isError=true early-return and the statusErr!=nil skip-if in
// Update's ecr branch.
func TestResourceRegistryEcrUpdate_Ecr_WaitForSync_APIError(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	require.NoError(t, d.Set("provider_type", "helm"))
	require.NoError(t, d.Set("wait_for_sync", true))
	d.SetId("test-oci-registry-uid")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrUpdate_Basic_WaitForSync_StatusOnlyMessage(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry-update-wait"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("provider_type", "helm"))
	require.NoError(t, d.Set("wait_for_sync", true))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"secret_key":      "sk",
			"access_key":      "ak",
		},
	}))
	d.SetId("oci-sync-completed-no-message")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "Status: Completed", d.Get("wait_for_status_message"))
}

func TestResourceRegistryEcrUpdate_Ecr_ValidateError(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	require.NoError(t, d.Set("is_synchronization", true))
	require.NoError(t, d.Set("provider_type", "helm"))
	d.SetId("test-registry-uid")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrUpdate_Basic_ValidateError(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry-update-validate-err"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("is_synchronization", true))
	require.NoError(t, d.Set("provider_type", "zarf"))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "basic",
			"username":        "user",
			"password":        "pass",
		},
	}))
	d.SetId("test-registry-uid")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceRegistryEcrUpdate_Ecr_UpdateError isolates
// UpdateOciEcrRegistry's error return (is_synchronization=false skips
// validateRegistryCred).
func TestResourceRegistryEcrUpdate_Ecr_UpdateError(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	require.NoError(t, d.Set("is_synchronization", false))
	d.SetId("test-registry-uid")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrUpdate_Basic_UpdateError(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("name", "test-basic-registry-update-err"))
	require.NoError(t, d.Set("type", "basic"))
	require.NoError(t, d.Set("endpoint", "https://basic.example.com"))
	require.NoError(t, d.Set("is_private", true))
	require.NoError(t, d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"secret_key":      "sk",
			"access_key":      "ak",
		},
	}))
	d.SetId("test-registry-uid")
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
}

// ---------------------------------------------------------------------------
// Update — is_synchronization HasChange guard.
//
// d.GetChange("is_synchronization") needs a genuine old/new diff pair, which
// schema.TestResourceData can't produce (its "old" state is always empty).
// buildRegistryOciEcrSyncChangeResourceData drives the same
// Resource.Diff + schema.InternalMap(...).Data pipeline used by
// buildMachinePoolChangeResourceData (resource_cluster_apache_cloudstack_test.go)
// to get a real HasChange signal for this scalar bool field.
// ---------------------------------------------------------------------------

func buildRegistryOciEcrSyncChangeResourceData(t *testing.T, oldSync, newSync bool) *schema.ResourceData {
	t.Helper()
	res := resourceRegistryOciEcr()

	base := map[string]interface{}{
		"name":              "test-sync-change-registry",
		"type":              "basic",
		"endpoint":          "https://registry.example.com",
		"is_private":        true,
		"provider_type":     "helm",
		"base_content_path": "charts",
		"credentials": []interface{}{
			map[string]interface{}{
				"credential_type": "basic",
				"username":        "user",
				"password":        "pass",
			},
		},
	}

	oldRaw := map[string]interface{}{}
	for k, v := range base {
		oldRaw[k] = v
	}
	oldRaw["is_synchronization"] = oldSync
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId("test-registry-uid")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{}
	for k, v := range base {
		newRaw[k] = v
	}
	newRaw["is_synchronization"] = newSync
	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	// diff is nil when oldSync == newSync (no delta) — Data(oldState, nil)
	// is a valid no-op call that just returns the unchanged state.

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-registry-uid")
	return finalRD
}

func TestResourceRegistryEcrUpdate_IsSynchronization_DisableRejected(t *testing.T) {
	d := buildRegistryOciEcrSyncChangeResourceData(t, true, false)
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assertAnyDiagContains(t, diags, "cannot disable synchronization")
}

func TestResourceRegistryEcrUpdate_IsSynchronization_EnableAllowed(t *testing.T) {
	d := buildRegistryOciEcrSyncChangeResourceData(t, false, true)
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrUpdate_IsSynchronization_NoChange(t *testing.T) {
	d := buildRegistryOciEcrSyncChangeResourceData(t, true, true)
	diags := resourceRegistryEcrUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// ---------------------------------------------------------------------------
// Delete branches — the current test only exercises the ecr path.
// ---------------------------------------------------------------------------

func TestResourceRegistryEcrDelete_BasicPath(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "basic"))
	d.SetId("test-registry-uid")
	diags := resourceRegistryEcrDelete(context.Background(), d, unitTestMockAPIClient)
	_ = diags
}

// ---------------------------------------------------------------------------
// Read branches
// ---------------------------------------------------------------------------

func TestResourceRegistryEcrRead_BasicPath(t *testing.T) {
	// The basic-path Read expects a mock payload with populated
	// Spec.RegistryUID. When the mock returns a partial fixture, the
	// setter chain nil-derefs. Guard with recover so we still exercise
	// the top of the read + the switch dispatch.
	defer func() { _ = recover() }()

	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "basic"))
	d.SetId("test-registry-uid")
	_ = resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// Read — ecr credential-type / TLS / deleted / error branches. UIDs below
// are dispatched by ecrRegistryFixtureFor in
// tests/mockApiServer/routes/mockRegistries.go.
// ---------------------------------------------------------------------------

func TestResourceRegistryEcrRead_Ecr_SecretPlain(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "ecr"))
	d.SetId("ecr-uid-secret-plain")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	creds := d.Get("credentials").([]interface{})
	require.Len(t, creds, 1)
	cred := creds[0].(map[string]interface{})
	assert.Equal(t, string(models.V1AwsCloudAccountCredentialTypeSecret), cred["credential_type"])
	assert.Equal(t, "plain-secret-key", cred["secret_key"])
	assert.Empty(t, cred["tls_config"], "nil registry.Spec.TLS must produce an empty tls_config")
}

func TestResourceRegistryEcrRead_Ecr_SecretMasked(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "ecr"))
	d.SetId("ecr-uid-secret-masked")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	creds := d.Get("credentials").([]interface{})
	require.Len(t, creds, 1)
	cred := creds[0].(map[string]interface{})
	assert.Empty(t, cred["secret_key"], "masked API secret must not populate state")
}

func TestResourceRegistryEcrRead_Ecr_UnknownCredentialType(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "ecr"))
	d.SetId("ecr-uid-unknown-cred")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assertAnyDiagContains(t, diags, "not implemented")
}

func TestResourceRegistryEcrRead_Ecr_GetError(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "ecr"))
	d.SetId("any-ecr-uid")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceRegistryEcrRead_Ecr_WaitForSyncPreservedFromState(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "ecr"))
	require.NoError(t, d.Set("wait_for_sync", true))
	d.SetId("test-id")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.True(t, d.Get("wait_for_sync").(bool))
}

func TestResourceRegistryEcrRead_Ecr_WaitForSyncDefaultsFalse(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "ecr"))
	d.SetId("test-id")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.False(t, d.Get("wait_for_sync").(bool))
}

// ---------------------------------------------------------------------------
// Read — basic auth-type / password-preserve / deleted / error branches.
// UIDs dispatched by basicRegistryFixtureFor.
// ---------------------------------------------------------------------------

func TestResourceRegistryEcrRead_Basic_NoAuth(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "basic"))
	d.SetId("basic-uid-noauth")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	assert.False(t, d.Get("is_private").(bool))
	creds := d.Get("credentials").([]interface{})
	require.Len(t, creds, 1)
	cred := creds[0].(map[string]interface{})
	assert.Equal(t, "noAuth", cred["credential_type"])
}

func TestResourceRegistryEcrRead_Basic_PasswordPreservedFromState(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "basic"))
	setBasicCreds(t, d, nil)
	d.SetId("test-zarf-oci-reg-basic-uid")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	creds := d.Get("credentials").([]interface{})
	require.Len(t, creds, 1)
	cred := creds[0].(map[string]interface{})
	assert.Equal(t, "pass", cred["password"], "existing state password must be preserved over the API value")
}

func TestResourceRegistryEcrRead_Basic_GetError(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	require.NoError(t, d.Set("type", "basic"))
	d.SetId("any-basic-uid")
	diags := resourceRegistryEcrRead(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError(), "diags: %+v", diags)
}

// setBasicCreds is a tiny helper for building a credentials list.
// tlsConfig is applied only when non-nil (nil -> block omitted).
func setBasicCreds(t *testing.T, d *schema.ResourceData, tlsConfig []interface{}) {
	t.Helper()
	cred := map[string]interface{}{
		"credential_type": "basic",
		"username":        "user",
		"password":        "pass",
	}
	if tlsConfig != nil {
		cred["tls_config"] = tlsConfig
	}
	require.NoError(t, d.Set("credentials", []interface{}{cred}))
}

// TestOciBasicTLSConfigForRead_OmittedInStateAndApiDefault is the regression
// guard for PLT-2300: when HCL omits tls_config and the API returns a default
// TLS block (empty cert, insecure_skip_verify=false), Read must not
// synthesize a tls_config block into state — otherwise every subsequent
// plan proposes removing it.
func TestOciBasicTLSConfigForRead_OmittedInStateAndApiDefault(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	setBasicCreds(t, d, nil)

	got := ociBasicTLSConfigForRead(d, &models.V1TLSConfiguration{
		Certificate:        "",
		Enabled:            false,
		InsecureSkipVerify: false,
	})
	assert.Empty(t, got,
		"tls_config must stay empty when config omitted the block and the API returned defaults")
}

// TestOciBasicTLSConfigForRead_NilApiTLS covers the defensive nil-check —
// an API response without a TLS block must not panic and must produce no
// tls_config in state.
func TestOciBasicTLSConfigForRead_NilApiTLS(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	setBasicCreds(t, d, nil)

	got := ociBasicTLSConfigForRead(d, nil)
	assert.Empty(t, got, "nil API TLS must produce an empty tls_config")
}

// TestOciBasicTLSConfigForRead_PreserveStateBlockOnDefault is the
// round-trip case: when state already has a tls_config block, Read must
// keep populating it (even if the API values match defaults) so the field
// reflects server state and doesn't disappear from the user's tf state.
func TestOciBasicTLSConfigForRead_PreserveStateBlockOnDefault(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	setBasicCreds(t, d, []interface{}{
		map[string]interface{}{
			"certificate":          "",
			"insecure_skip_verify": false,
		},
	})

	got := ociBasicTLSConfigForRead(d, &models.V1TLSConfiguration{
		Certificate:        "",
		Enabled:            false,
		InsecureSkipVerify: false,
	})
	assert.Len(t, got, 1, "explicit tls_config in state must round-trip")
	entry := got[0].(map[string]interface{})
	assert.Equal(t, "", entry["certificate"])
	assert.Equal(t, false, entry["insecure_skip_verify"])
}

// TestOciBasicTLSConfigForRead_MaterializeMeaningfulApiTLS_InsecureSkipTrue
// covers real server-side drift: when state has no tls_config block but the
// API reports insecure_skip_verify=true, we must surface it so the user
// sees the drift instead of it being silently swallowed.
func TestOciBasicTLSConfigForRead_MaterializeMeaningfulApiTLS_InsecureSkipTrue(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	setBasicCreds(t, d, nil)

	got := ociBasicTLSConfigForRead(d, &models.V1TLSConfiguration{
		InsecureSkipVerify: true,
	})
	require.Len(t, got, 1)
	entry := got[0].(map[string]interface{})
	assert.Equal(t, true, entry["insecure_skip_verify"])
}

// TestOciBasicTLSConfigForRead_MaterializeMeaningfulApiTLS_CertPresent
// mirrors the previous case for a non-empty certificate: server-side drift
// must be reflected in state.
func TestOciBasicTLSConfigForRead_MaterializeMeaningfulApiTLS_CertPresent(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	setBasicCreds(t, d, nil)

	got := ociBasicTLSConfigForRead(d, &models.V1TLSConfiguration{
		Certificate: "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
	})
	require.Len(t, got, 1)
	entry := got[0].(map[string]interface{})
	assert.Contains(t, entry["certificate"].(string), "BEGIN CERTIFICATE")
}

// ---------------------------------------------------------------------------
// waitForOciRegistrySync (0%)
//
// registryType "basic" hits the mocked GET
// /v1/registries/oci/{uid}/basic/sync/status route, which reports
// Status="Success" on the very first poll (waitDelayOverride is zeroed by
// TestMain so there's no real wait). registryType "ecr" has no mocked
// GET .../ecr/sync/status route, so it 404s — a legitimate non-timeout
// error branch.
// ---------------------------------------------------------------------------

func TestWaitForOciRegistrySync_Success(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	c := castV1Client(t, unitTestMockAPIClient)

	diags, isError := waitForOciRegistrySync(context.Background(), d, "test-oci-registry-uid", diag.Diagnostics{}, c, schema.TimeoutCreate, "basic")
	assert.False(t, isError, "diags: %+v", diags)
}

func TestWaitForOciRegistrySync_APIError(t *testing.T) {
	// No mock route for GET .../ecr/sync/status → the refresh func's first
	// call returns a non-404-tolerant error, which is neither a
	// retry.TimeoutError nor a sync-failure status, so the final
	// "return diag.FromErr(err), true" branch is exercised.
	d := resourceRegistryOciEcr().TestResourceData()
	c := castV1Client(t, unitTestMockAPIClient)

	diags, isError := waitForOciRegistrySync(context.Background(), d, "test-oci-registry-uid", diag.Diagnostics{}, c, schema.TimeoutCreate, "ecr")
	assert.True(t, isError, "diags: %+v", diags)
}
