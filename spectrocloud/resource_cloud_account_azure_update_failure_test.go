package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PLT-2558: when Palette rejects an update, state must keep the values from before the apply. Otherwise the rejected
// values are saved, and because Palette never returns azure_client_secret, refresh cannot correct it: the next plan
// shows no changes even though Palette still has the old secret.
func TestCloudAccountAzureUpdateFailureKeepsPriorState(t *testing.T) {
	ctx := context.Background()
	res := resourceCloudAccountAzure()

	prior := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                "azure-account",
		"context":             "project",
		"azure_tenant_id":     "tenant-id",
		"azure_client_id":     "client-id-old",
		"azure_client_secret": "secret-old",
	})
	prior.SetId("azure-account-uid")
	state := prior.State()
	require.NotNil(t, state)

	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"name":                "azure-account",
		"context":             "project",
		"azure_tenant_id":     "tenant-id",
		"azure_client_id":     "client-id-new",
		"azure_client_secret": "secret-new",
	})
	diff, err := res.Diff(ctx, state, config, unitTestMockAPINegativeClient)
	require.NoError(t, err)
	require.NotNil(t, diff, "precondition: the plan has changes to apply")

	// The negative mock API rejects the update.
	newState, diags := res.Apply(ctx, state, diff, unitTestMockAPINegativeClient)
	require.True(t, diags.HasError(), "precondition: the update must fail")
	require.NotNil(t, newState)

	assert.Equal(t, "secret-old", newState.Attributes["azure_client_secret"])
	assert.Equal(t, "client-id-old", newState.Attributes["azure_client_id"])
	assert.Equal(t, "azure-account-uid", newState.ID)
}

// A successful update still saves the new values.
func TestCloudAccountAzureUpdateSuccessSavesNewState(t *testing.T) {
	ctx := context.Background()
	res := resourceCloudAccountAzure()

	prior := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                "azure-account",
		"context":             "project",
		"azure_tenant_id":     "tenant-id",
		"azure_client_id":     "client-id",
		"azure_client_secret": "secret-old",
	})
	prior.SetId("azure-account-uid")
	state := prior.State()

	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"name":                "azure-account",
		"context":             "project",
		"azure_tenant_id":     "tenant-id",
		"azure_client_id":     "client-id",
		"azure_client_secret": "secret-new",
	})
	diff, err := res.Diff(ctx, state, config, unitTestMockAPIClient)
	require.NoError(t, err)

	newState, diags := res.Apply(ctx, state, diff, unitTestMockAPIClient)
	require.False(t, diags.HasError(), "update against the mock API must succeed: %v", diags)
	assert.Equal(t, "secret-new", newState.Attributes["azure_client_secret"])
}
