package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
