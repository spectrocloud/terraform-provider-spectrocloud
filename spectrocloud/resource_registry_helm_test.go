package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareResourceRegistryHelm() *schema.ResourceData {
	d := resourceRegistryHelm().TestResourceData()
	d.SetId("test-reg-id")
	_ = d.Set("name", "test-reg-name")
	_ = d.Set("is_private", true)
	_ = d.Set("endpoint", "test.com")
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

func TestResourceRegistryHelmCreate(t *testing.T) {
	d := prepareResourceRegistryHelm()
	var diags diag.Diagnostics
	var ctx context.Context
	diags = resourceRegistryHelmCreate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

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

func TestResourceRegistryHelmRead(t *testing.T) {
	d := prepareResourceRegistryHelm()
	var diags diag.Diagnostics
	var ctx context.Context
	diags = resourceRegistryHelmRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestResourceRegistryHelmUpdate(t *testing.T) {
	d := prepareResourceRegistryHelm()
	var diags diag.Diagnostics
	var ctx context.Context
	diags = resourceRegistryHelmUpdate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestResourceRegistryHelmDelete(t *testing.T) {
	d := prepareResourceRegistryHelm()
	var diags diag.Diagnostics
	var ctx context.Context
	diags = resourceRegistryHelmDelete(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
