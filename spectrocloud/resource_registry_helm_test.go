package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func prepareResourceRegistryHelm() *schema.ResourceData {
	d := resourceRegistryHelm().TestResourceData()
	// d.SetId("test-reg-id")
	_ = d.Set("name", "test-reg-name")
	_ = d.Set("is_private", true)
	_ = d.Set("endpoint", "test.com")
	_ = d.Set("wait_for_sync", false)
	var cred []interface{}
	cred = append(cred, map[string]interface{}{
		"credential_type": "token",
		"username":        "test-username",
		"password":        "test-password",
		"token":           "test_token",
	})
	_ = d.Set("credentials", cred)
	return d
}

func TestResourceRegistryHelmCRUD(t *testing.T) {
	testResourceCRUD(t, prepareResourceRegistryHelm, unitTestMockAPIClient,
		resourceRegistryHelmCreate, resourceRegistryHelmRead, resourceRegistryHelmUpdate, resourceRegistryHelmDelete)
}

// func TestResourceRegistryHelmCreate(t *testing.T) {
// 	d := prepareResourceRegistryHelm()
// 	var diags diag.Diagnostics
// 	var ctx context.Context
// 	diags = resourceRegistryHelmCreate(ctx, d, unitTestMockAPIClient)
// 	assert.Equal(t, 0, len(diags))
// }

func TestResourceRegistryHelmCreateNoAuth(t *testing.T) {
	d := prepareResourceRegistryHelm()
	var diags diag.Diagnostics
	var ctx context.Context
	var cred []interface{}
	cred = append(cred, map[string]interface{}{
		"credential_type": "noAuth",
		"username":        "test-username",
		"password":        "test-password",
		"token":           "test_token",
	})
	_ = d.Set("credentials", cred)
	diags = resourceRegistryHelmCreate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestResourceRegistryHelmCreateBasic(t *testing.T) {
	d := prepareResourceRegistryHelm()
	var diags diag.Diagnostics
	var ctx context.Context
	var cred []interface{}
	cred = append(cred, map[string]interface{}{
		"credential_type": "basic",
		"username":        "test-username",
		"password":        "test-password",
		"token":           "test_token",
	})
	_ = d.Set("credentials", cred)
	diags = resourceRegistryHelmCreate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestResourceRegistryHelmCreateWithWaitForSync(t *testing.T) {
	d := prepareResourceRegistryHelm()
	_ = d.Set("wait_for_sync", true)
	var diags diag.Diagnostics
	ctx := context.Background()
	diags = resourceRegistryHelmCreate(ctx, d, unitTestMockAPIClient)
	// Should complete successfully with no errors or warnings
	assert.Equal(t, 0, len(diags))
}

func TestResourceRegistryHelmUpdateWithWaitForSync(t *testing.T) {
	d := prepareResourceRegistryHelm()
	d.SetId("test-registry-uid") // Update and wait_for_sync require an existing resource ID (mock uses this UID)
	_ = d.Set("wait_for_sync", true)
	var diags diag.Diagnostics
	ctx := context.Background()
	diags = resourceRegistryHelmUpdate(ctx, d, unitTestMockAPIClient)
	// Should complete successfully with no errors or warnings
	assert.Equal(t, 0, len(diags))
}

// The mock's helm registry payload returns password "test=pwd" and token "as",
// which differ from what's set in state below. Read must preserve the
// state values rather than overwriting them with the API's response, or
// every subsequent plan reports spurious drift (PLT-2400).
func TestResourceRegistryHelmReadPreservesCredentialsFromState(t *testing.T) {
	d := prepareResourceRegistryHelm()
	d.SetId("test-registry-uid")
	var cred []interface{}
	cred = append(cred, map[string]interface{}{
		"credential_type": "token",
		"username":        "test-username",
		"password":        "",
		"token":           "state-token",
	})
	_ = d.Set("credentials", cred)

	var diags diag.Diagnostics
	ctx := context.Background()
	diags = resourceRegistryHelmRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))

	creds := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "state-token", creds["token"])
}

func TestResourceRegistryHelmReadPreservesPasswordFromState(t *testing.T) {
	d := prepareResourceRegistryHelm()
	d.SetId("test-helm-basic-uid")
	var cred []interface{}
	cred = append(cred, map[string]interface{}{
		"credential_type": "basic",
		"username":        "test-username",
		"password":        "state-password",
		"token":           "",
	})
	_ = d.Set("credentials", cred)

	var diags diag.Diagnostics
	ctx := context.Background()
	diags = resourceRegistryHelmRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))

	creds := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "state-password", creds["password"])
}

// PLT-2356: spectrocloud_registry_helm had no TLS knob and
// toRegistryHelmCredential never populated V1RegistryAuth.TLS, so custom CA,
// mTLS and insecureSkipVerify were unreachable from Terraform. The block is
// nested under credentials as tls_config, matching spectrocloud_registry_oci
// and the API's spec.auth.tls shape, and is sent only when configured.

func helmCredentialsWithTLS(tls []interface{}) []interface{} {
	cred := map[string]interface{}{
		"credential_type": "basic",
		"username":        "test-username",
		"password":        "test-password",
	}
	if tls != nil {
		cred["tls_config"] = tls
	}
	return []interface{}{cred}
}

func TestPLT2356_HelmRegistryTLSSchema(t *testing.T) {
	creds := resourceRegistryHelm().Schema["credentials"].Elem.(*schema.Resource).Schema
	tlsSchema, ok := creds["tls_config"]
	assert.True(t, ok, "expected a tls_config block inside credentials")
	assert.Equal(t, schema.TypeList, tlsSchema.Type)
	assert.True(t, tlsSchema.Optional)
	assert.Equal(t, 1, tlsSchema.MaxItems)

	_, topLevel := resourceRegistryHelm().Schema["tls"]
	assert.False(t, topLevel, "tls must live under credentials, not at the top level")

	fields := tlsSchema.Elem.(*schema.Resource).Schema

	tests := []struct {
		name        string
		field       string
		wantType    schema.ValueType
		wantDefault interface{}
		sensitive   bool
	}{
		{"enabled defaults to true", "enabled", schema.TypeBool, true, false},
		{"ca is a plain string", "ca", schema.TypeString, nil, false},
		{"certificate is a plain string", "certificate", schema.TypeString, nil, false},
		{"key is sensitive", "key", schema.TypeString, nil, true},
		{"insecure_skip_verify defaults to false", "insecure_skip_verify", schema.TypeBool, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ok := fields[tt.field]
			assert.True(t, ok, "missing tls_config field %q", tt.field)
			assert.Equal(t, tt.wantType, f.Type)
			assert.True(t, f.Optional)
			assert.Equal(t, tt.wantDefault, f.Default)
			assert.Equal(t, tt.sensitive, f.Sensitive)
		})
	}
}

func TestPLT2356_HelmRegistryTLSPayload(t *testing.T) {
	tests := []struct {
		name string
		tls  []interface{}
		want *models.V1TLSConfiguration
	}{
		{
			name: "omitted block sends no tls",
			tls:  nil,
			want: nil,
		},
		{
			name: "enabled with ca only",
			tls: []interface{}{map[string]interface{}{
				"enabled": true, "ca": "ca-pem", "certificate": "", "key": "", "insecure_skip_verify": false,
			}},
			want: &models.V1TLSConfiguration{Enabled: true, Ca: "ca-pem"},
		},
		{
			name: "mutual tls with client certificate",
			tls: []interface{}{map[string]interface{}{
				"enabled": true, "ca": "ca-pem", "certificate": "cert-pem", "key": "key-pem", "insecure_skip_verify": false,
			}},
			want: &models.V1TLSConfiguration{Enabled: true, Ca: "ca-pem", Certificate: "cert-pem", Key: "key-pem"},
		},
		{
			name: "insecure skip verify",
			tls: []interface{}{map[string]interface{}{
				"enabled": true, "ca": "", "certificate": "", "key": "", "insecure_skip_verify": true,
			}},
			want: &models.V1TLSConfiguration{Enabled: true, InsecureSkipVerify: true},
		},
		{
			name: "explicitly disabled",
			tls: []interface{}{map[string]interface{}{
				"enabled": false, "ca": "", "certificate": "", "key": "", "insecure_skip_verify": false,
			}},
			want: &models.V1TLSConfiguration{Enabled: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := prepareResourceRegistryHelm()
			_ = d.Set("credentials", helmCredentialsWithTLS(tt.tls))
			// Create and update build separate payload types from the same block.
			assert.Equal(t, tt.want, toRegistryEntityHelm(d).Spec.Auth.TLS, "create path")
			assert.Equal(t, tt.want, toRegistryHelm(d).Spec.Auth.TLS, "update path")
		})
	}
}

// Read emits tls_config only when the practitioner configured one or the API
// returned something meaningful, so registries stored without TLS show no diff.
func TestPLT2356_HelmRegistryTLSRead(t *testing.T) {
	tests := []struct {
		name       string
		configured bool
		want       []interface{}
	}{
		{
			name:       "configured block round-trips from the api",
			configured: true,
			want: []interface{}{map[string]interface{}{
				"enabled": true, "ca": "test-ca-pem", "certificate": "test-cert-pem",
				"key": "test-key-pem", "insecure_skip_verify": true,
			}},
		},
		{
			name:       "unconfigured block still surfaces meaningful api tls",
			configured: false,
			want: []interface{}{map[string]interface{}{
				"enabled": true, "ca": "test-ca-pem", "certificate": "test-cert-pem",
				"key": "test-key-pem", "insecure_skip_verify": true,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := resourceRegistryHelm().TestResourceData()
			d.SetId("helm-uid-tls")
			if tt.configured {
				_ = d.Set("credentials", helmCredentialsWithTLS([]interface{}{
					map[string]interface{}{"enabled": true},
				}))
			}

			diags := resourceRegistryHelmRead(context.Background(), d, unitTestMockAPIClient)
			assert.False(t, diags.HasError(), "unexpected error: %v", diags)

			creds := d.Get("credentials").([]interface{})
			assert.Len(t, creds, 1)
			assert.Equal(t, tt.want, creds[0].(map[string]interface{})["tls_config"])
		})
	}
}

// A registry whose API payload carries no meaningful TLS emits no tls_config.
func TestPLT2356_HelmRegistryTLSReadOmittedWhenEmpty(t *testing.T) {
	d := resourceRegistryHelm().TestResourceData()
	d.SetId("test-registry-uid")

	diags := resourceRegistryHelmRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "unexpected error: %v", diags)

	creds := d.Get("credentials").([]interface{})
	assert.Len(t, creds, 1)
	assert.Empty(t, creds[0].(map[string]interface{})["tls_config"])
}

// PLT-2401: spectrocloud_registry_helm had only `is_private`, unlike
// spectrocloud_registry_oci's `is_private` + `is_synchronization` pair. The
// Helm API has no independent sync flag, so `is_synchronization` is added as
// a naming-parity alias over the same underlying value, `is_private` is
// deprecated, and the two are mutually exclusive via ExactlyOneOf.

func TestPLT2401_HelmRegistryIsSynchronizationSchema(t *testing.T) {
	s := resourceRegistryHelm().Schema

	isPrivate, ok := s["is_private"]
	assert.True(t, ok, "expected an is_private attribute")
	assert.True(t, isPrivate.Optional)
	assert.True(t, isPrivate.Computed)
	assert.NotEmpty(t, isPrivate.Deprecated, "is_private should be marked deprecated")
	assert.ElementsMatch(t, []string{"is_private", "is_synchronization"}, isPrivate.ExactlyOneOf)

	isSync, ok := s["is_synchronization"]
	assert.True(t, ok, "expected an is_synchronization attribute")
	assert.Equal(t, schema.TypeBool, isSync.Type)
	assert.True(t, isSync.Optional)
	assert.True(t, isSync.Computed)
	assert.Empty(t, isSync.Deprecated)
	assert.ElementsMatch(t, []string{"is_private", "is_synchronization"}, isSync.ExactlyOneOf)
}

func TestPLT2401_ResolveHelmIsPrivate(t *testing.T) {
	tests := []struct {
		name         string
		isPrivate    bool
		isSync       bool
		wantResolved bool
	}{
		{"only is_private true", true, false, true},
		{"only is_private false", false, false, false},
		{"only is_synchronization true", false, true, true},
		{"only is_synchronization false", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := prepareResourceRegistryHelm()
			_ = d.Set("is_private", tt.isPrivate)
			_ = d.Set("is_synchronization", tt.isSync)
			assert.Equal(t, tt.wantResolved, resolveHelmIsPrivate(d))
			assert.Equal(t, tt.wantResolved, toRegistryEntityHelm(d).Spec.IsPrivate)
			assert.Equal(t, tt.wantResolved, toRegistryHelm(d).Spec.IsPrivate)
		})
	}
}

// Read must mirror the single API isPrivate value into both is_private and
// is_synchronization so neither attribute drifts regardless of which one the
// practitioner configured.
func TestPLT2401_HelmRegistryReadMirrorsIsPrivateAndIsSynchronization(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want bool
	}{
		{"private registry", "test-helm-private-uid", true},
		{"public registry", "test-registry-uid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := resourceRegistryHelm().TestResourceData()
			d.SetId(tt.uid)

			diags := resourceRegistryHelmRead(context.Background(), d, unitTestMockAPIClient)
			assert.False(t, diags.HasError(), "unexpected error: %v", diags)

			assert.Equal(t, tt.want, d.Get("is_private"))
			assert.Equal(t, tt.want, d.Get("is_synchronization"))
		})
	}
}
